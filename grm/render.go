package grm

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
	"unsafe"
)

var batchSize = 16384

const drawCount = 1

const vec4Size = 16

const vertSize = 12

type UniformBufferObject struct {
	obj  mgl32.Mat4
	cam  mgl32.Mat4
	proj mgl32.Mat4
	amb mgl32.Vec4
}

type Renderer struct {
	vki      *vk.Instance
	physical *vk.PhysicalDevice
	device   *vk.Device

	gfxq *vk.Queue     // graphics queue
	pq   *vk.Queue     // presentation queue
	sc   *vk.Swapchain // swap chain

	sci  []vk.Image  // swap chain images
	scif vk.Format   // image format
	sce  vk.Extent2D // swap chain extent

	sciv []vk.ImageView // swap chain views

	pipeline       vk.Pipeline // graphics pipeline
	pipelineCache  *vk.PipelineCache
	renderPass     *vk.RenderPass
	pipelineLayout *vk.PipelineLayout // layout

	descriptorLayout *vk.DescriptorSetLayout
	descriptorPool   *vk.DescriptorPool
	descriptorSets   *vk.DescriptorSet

	framebuffers []vk.Framebuffer

	commandPool *vk.CommandPool
	commands    []vk.CommandBuffer

	available []*vk.Semaphore // image is available
	finished  []*vk.Semaphore // rendering is finished

	inf []*vk.Fence // fences in flight

	maxFrames int // max frames in flight
	frame     int // current frame

	ticks float64

	validation       []string // validation layers
	enableValidation bool

	surface vk.Surface

	dextensions []string // device extensions
	extensions  []string // list of extensions

	window *glfw.Window

	width  int
	height int

	setCount uint32

	drawCount int // count per frame
	draws int // current draw count

	depthImage *vk.Image
	depthImageMem *vk.DeviceMemory
	depthImageView *vk.ImageView

	vertexBuffer    *vk.Buffer
	vertexBufferMem *vk.DeviceMemory
	indexBuffer     *vk.Buffer
	indexBufferMem  *vk.DeviceMemory

	drawBuffer    *vk.Buffer
	drawBufferMem *vk.DeviceMemory

	uniformBuffers   *vk.Buffer
	uniformBufferMem *vk.DeviceMemory

	texture *Texture

	ubo UniformBufferObject

	vertexData []float32
	indexData  []uint16

	Ambient mgl32.Vec4
}

func (r *Renderer) init(w *glfw.Window) {
	r.window = w

	r.enableValidation = false

	r.frame = 0
	r.maxFrames = 2
	r.setCount = 1

	r.width = 1600
	r.height = 900

	r.ubo = UniformBufferObject{}

	r.ubo.obj = mgl32.Ident4()
	r.ubo.cam = mgl32.Ident4()
	r.ubo.proj = mgl32.Ident4()

	r.vkInit() // init vulkan
}

// VULKAN BOILERPLATE
func (r *Renderer) vkInit() {
	var err error
	err = vk.Init() // initialize Vulkan
	if err != nil {
		panic(err)
	}

	r.dextensions = []string{
		string([]byte(vk.KhrSwapchainExtensionName)),
	}

	r.extensions = r.window.GetRequiredInstanceExtensions()

	r.createInstance()

	if r.enableValidation {
		r.validation = validationLayers()
	}
	r.validation = nil

	r.getPhysical() // select physical gpu

	var properties vk.PhysicalDeviceProperties

	vk.GetPhysicalDeviceProperties(*r.physical, &properties)
	properties.Deref() // deref to read c values
	properties.Limits.Deref()

	maxbuf := properties.Limits.MaxStorageBufferRange

	println("max buffer: ", maxbuf)

	if maxbuf > uint32(batchSize) {
		batchSize = int(maxbuf / 8)

		if batchSize > 1024*1000*10 {
			batchSize = 1024*1000*10
		}
	}

	//r.dextensions = r.getExtensions() // just get all of them

	// SURFACE CREATION
	var sup uintptr
	sup, err = r.window.CreateWindowSurface(*r.vki, nil)
	if err != nil {
		panic(err)
	}
	r.surface = vk.SurfaceFromPointer(sup)

	r.createLogical() // interface

	r.createSwap() // swap chain creation

	r.createViews() // views for swap chain

	r.createRenderPass()

	r.createDescriptor()

	r.createPipeline() // graphics pipeline

	r.createCommandPool()

	r.createDepth()

	r.createFramebuffers()

	r.createVertexBuffer()

	r.createIndexBuffer()

	r.createUniformBuffers()

	r.createDescriptorPool()

	r.texture = new(Texture)
	r.texture.Load(r, "gfx/textures.png")

	r.createDescriptorSets()

	r.createCommandBuffers()

	r.createSync()
}

func (r *Renderer) updateTex() {

	r.createDescriptorSets()
	r.updateCommandBuffers()
}

func (r *Renderer) createDepth() {
	format := vk.FormatD32Sfloat // finddepthformat?

	r.depthImage = new(vk.Image)
	r.depthImageView = new(vk.ImageView)
	r.depthImageMem = new(vk.DeviceMemory)
	r.createImage(r.sce.Width, r.sce.Height, format, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit), r.depthImage, r.depthImageMem)

	view := r.createImageView(*r.depthImage, format, vk.ImageAspectFlags(vk.ImageAspectDepthBit))
	r.depthImageView = &view

	r.transitionLayout(*r.depthImage, format, vk.ImageLayoutUndefined, vk.ImageLayoutDepthStencilAttachmentOptimal)
}

func (r *Renderer) createImageView(image vk.Image, format vk.Format, flags vk.ImageAspectFlags) vk.ImageView {
	info := vk.ImageViewCreateInfo{}
	info.SType = vk.StructureTypeImageViewCreateInfo
	info.Image = image
	info.ViewType = vk.ImageViewType2d
	info.Format = format
	info.SubresourceRange.AspectMask = flags
	info.SubresourceRange.BaseMipLevel = 0
	info.SubresourceRange.LevelCount = 1
	info.SubresourceRange.BaseArrayLayer = 0
	info.SubresourceRange.LayerCount = 1

	var view vk.ImageView

	vkr := vk.CreateImageView(*r.device, &info, nil, &view)

	if vkr != vk.Success {
		panic(vkr)
	}

	return view
}

func (r *Renderer) createDescriptorSets() {
	layouts := make([]vk.DescriptorSetLayout, r.setCount)

	for i := 0; i < len(layouts); i++ {
		layouts[i] = *r.descriptorLayout
	}

	allocInfo := vk.DescriptorSetAllocateInfo{}
	allocInfo.SType = vk.StructureTypeDescriptorSetAllocateInfo
	allocInfo.DescriptorPool = *r.descriptorPool
	allocInfo.DescriptorSetCount = r.setCount
	allocInfo.PSetLayouts = layouts

	//r.descriptorSets = make([]*vk.DescriptorSet, len(r.sci))

	//for i := 0; i < len(r.descriptorSets); i++ {
	r.descriptorSets = new(vk.DescriptorSet)
	vkr := vk.AllocateDescriptorSets(*r.device, &allocInfo, r.descriptorSets)

	if vkr != vk.Success {
		panic(vkr)
	}
	//}

	//for i := 0; i < r.setCount; i++ {
	bufInfo := vk.DescriptorBufferInfo{}
	bufInfo.Buffer = *r.uniformBuffers
	bufInfo.Offset = 0
	bufInfo.Range = vk.DeviceSize(r.getUBOSize())

	imageInfo := vk.DescriptorImageInfo{}
	imageInfo.ImageLayout = vk.ImageLayoutShaderReadOnlyOptimal
	imageInfo.ImageView = *r.texture.imageView
	imageInfo.Sampler = *r.texture.sampler

	write := make([]vk.WriteDescriptorSet, 2)
	write[0].SType = vk.StructureTypeWriteDescriptorSet
	write[0].DstSet = *r.descriptorSets
	write[0].DstBinding = 0
	write[0].DstArrayElement = 0
	write[0].DescriptorType = vk.DescriptorTypeUniformBuffer
	write[0].DescriptorCount = 1
	write[0].PBufferInfo = []vk.DescriptorBufferInfo{bufInfo}
	write[0].PImageInfo = nil
	write[0].PTexelBufferView = nil

	write[1].SType = vk.StructureTypeWriteDescriptorSet
	write[1].DstSet = *r.descriptorSets
	write[1].DstBinding = 1
	write[1].DstArrayElement = 0
	write[1].DescriptorType = vk.DescriptorTypeCombinedImageSampler
	write[1].DescriptorCount = 1
	write[1].PImageInfo = []vk.DescriptorImageInfo{imageInfo}

	vk.UpdateDescriptorSets(*r.device, uint32(len(write)), write, 0, nil)
	//}
}

func (r *Renderer) createDescriptorPool() {
	size := make([]vk.DescriptorPoolSize, 2)
	size[0].Type = vk.DescriptorTypeUniformBuffer
	size[0].DescriptorCount = r.setCount
	size[1].Type = vk.DescriptorTypeCombinedImageSampler
	size[1].DescriptorCount = r.setCount

	info := vk.DescriptorPoolCreateInfo{}
	info.SType = vk.StructureTypeDescriptorPoolCreateInfo
	info.PoolSizeCount = uint32(len(size))
	info.PPoolSizes = size
	info.MaxSets = r.setCount

	r.descriptorPool = new(vk.DescriptorPool)
	vkr := vk.CreateDescriptorPool(*r.device, &info, nil, r.descriptorPool)

	if vkr != vk.Success {
		panic(vkr)
	}
}

func (r *Renderer) createUniformBuffers() {
	bSize := r.getUBOSize()
	//size := len(r.sci)

	//r.uniformBuffers = make([]*vk.Buffer, size)
	//r.uniformBufferMem = make([]*vk.DeviceMemory, size)
	r.uniformBuffers = new(vk.Buffer)
	r.uniformBufferMem = new(vk.DeviceMemory)

	//for i := 0; i < size; i++ {
	//	r.uniformBuffers[i] = new(vk.Buffer)
	//	r.uniformBufferMem[i] = new(vk.DeviceMemory)
	r.createBuffer(vk.DeviceSize(bSize), vk.BufferUsageFlags(vk.BufferUsageUniformBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
		r.uniformBuffers, r.uniformBufferMem)
	//}
}

func (r *Renderer) createDescriptor() {
	uboBinding := vk.DescriptorSetLayoutBinding{}

	uboBinding.Binding = 0
	uboBinding.DescriptorType = vk.DescriptorTypeUniformBuffer
	uboBinding.DescriptorCount = 1
	uboBinding.StageFlags = vk.ShaderStageFlags(vk.ShaderStageVertexBit)
	uboBinding.PImmutableSamplers = nil

	samplerBinding := vk.DescriptorSetLayoutBinding{}
	samplerBinding.Binding = 1
	samplerBinding.DescriptorCount = 1
	samplerBinding.DescriptorType = vk.DescriptorTypeCombinedImageSampler
	samplerBinding.PImmutableSamplers = nil
	samplerBinding.StageFlags = vk.ShaderStageFlags(vk.ShaderStageFragmentBit)

	bindings := []vk.DescriptorSetLayoutBinding{uboBinding, samplerBinding}
	info := vk.DescriptorSetLayoutCreateInfo{}
	info.SType = vk.StructureTypeDescriptorSetLayoutCreateInfo
	info.BindingCount = uint32(len(bindings))
	info.PBindings = bindings

	r.descriptorLayout = new(vk.DescriptorSetLayout)
	vkr := vk.CreateDescriptorSetLayout(*r.device, &info, nil, r.descriptorLayout)

	if vkr != vk.Success {
		panic(vkr)
	}
}

func (r *Renderer) createIndexBuffer() {
	bSize := vk.DeviceSize(getIntSize(r.indexData))

	if bSize == 0 {
		bSize = 1
	}

	bSize = vk.DeviceSize(uint32(batchSize))
	/*
		var stagingBuffer *vk.Buffer
		var stagingBufferMem *vk.DeviceMemory
		stagingBuffer = new(vk.Buffer)
		stagingBufferMem = new(vk.DeviceMemory)

		r.createBuffer(bSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
			vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
			stagingBuffer, stagingBufferMem)

		iData := intData(r.indexData)

		var data unsafe.Pointer
		vk.MapMemory(*r.device, *stagingBufferMem, 0, bSize, 0, &data)
		vk.Memcopy(data, iData)
		vk.UnmapMemory(*r.device, *stagingBufferMem)
	*/
	r.indexBuffer = new(vk.Buffer)
	r.indexBufferMem = new(vk.DeviceMemory)

	r.createBuffer(bSize, vk.BufferUsageFlags(vk.BufferUsageIndexBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
		r.indexBuffer, r.indexBufferMem)
	/*
		r.copyBuffer(stagingBuffer, r.indexBuffer, bSize)

		vk.DestroyBuffer(*r.device, *stagingBuffer, nil)
		vk.FreeMemory(*r.device, *stagingBufferMem, nil)
		stagingBuffer = nil
		stagingBufferMem = nil*/
	r.updateIndex()
}

func (r *Renderer) updateIndex() {
	bSize := vk.DeviceSize(getFloatSize(r.vertexData))
	if bSize == 0 {
		bSize = 1
	}

	iData := intData(r.indexData)

	var data unsafe.Pointer
	vk.MapMemory(*r.device, *r.indexBufferMem, 0, bSize, 0, &data)
	vk.Memcopy(data, iData)
	vk.UnmapMemory(*r.device, *r.indexBufferMem)

}

func (r *Renderer) updateVertex() {
	bSize := vk.DeviceSize(getFloatSize(r.vertexData))

	if bSize == 0 {
		bSize = 1
	}

	if bSize >= vk.DeviceSize(uint32(batchSize)) {
		r.Clear()
		return
	}
	/*
		var stagingBuffer *vk.Buffer
		var stagingBufferMem *vk.DeviceMemory
		stagingBuffer = new(vk.Buffer)
		stagingBufferMem = new(vk.DeviceMemory)

		r.createBuffer(bSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
			vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
			stagingBuffer, stagingBufferMem)

		vData := floatData(r.vertexData)

		var data unsafe.Pointer
		vk.MapMemory(*r.device, *stagingBufferMem, 0, bSize, 0, &data)
		vk.Memcopy(data, vData)
		vk.UnmapMemory(*r.device, *stagingBufferMem)
		data = nil

		r.copyBuffer(stagingBuffer, r.vertexBuffer, bSize)

		vk.DestroyBuffer(*r.device, *stagingBuffer, nil)
		vk.FreeMemory(*r.device, *stagingBufferMem, nil)

		stagingBuffer = nil
		stagingBufferMem = nil*/

	vData := floatData(r.vertexData)

	var data unsafe.Pointer
	vk.MapMemory(*r.device, *r.vertexBufferMem, 0, bSize, 0, &data)
	vk.Memcopy(data, vData)
	vk.UnmapMemory(*r.device, *r.vertexBufferMem)
	data = nil

}

func drawData(dat []vk.DrawIndexedIndirectCommand) []byte {
	const m = 0x7fffffff
	s := int(unsafe.Sizeof(dat[0]))
	return (*[m]byte)(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&dat)).data))[:len(dat)*s]
}

func (r *Renderer) updateDraw() {
	bSize := vk.DeviceSize(uint32(unsafe.Sizeof(vk.DrawIndexedIndirectCommand{})))

	indirect := vk.DrawIndexedIndirectCommand{}
	indirect.InstanceCount = 1
	indirect.FirstInstance = 0
	indirect.FirstIndex = 0
	indirect.IndexCount = uint32(len(r.indexData))

	dData := drawData([]vk.DrawIndexedIndirectCommand{indirect})

	var data unsafe.Pointer
	vk.MapMemory(*r.device, *r.drawBufferMem, 0, bSize, 0, &data)
	vk.Memcopy(data, dData)
	vk.UnmapMemory(*r.device, *r.drawBufferMem)
	data = nil
}

func (r *Renderer) createVertexBuffer() {
	bSize := vk.DeviceSize(getFloatSize(r.vertexData))

	if bSize == 0 {
		bSize = 1
	}

	bSize = vk.DeviceSize(uint32(batchSize))

	/*
		var stagingBuffer *vk.Buffer
		var stagingBufferMem *vk.DeviceMemory
		stagingBuffer = new(vk.Buffer)
		stagingBufferMem = new(vk.DeviceMemory)

		r.createBuffer(bSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
			vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
			stagingBuffer, stagingBufferMem)

		vData := floatData(r.vertexData)

		var data unsafe.Pointer
		vk.MapMemory(*r.device, *stagingBufferMem, 0, bSize, 0, &data)
		vk.Memcopy(data, vData)
		vk.UnmapMemory(*r.device, *stagingBufferMem)
	*/
	r.vertexBuffer = new(vk.Buffer)
	r.vertexBufferMem = new(vk.DeviceMemory)

	r.createBuffer(bSize, vk.BufferUsageFlags(vk.BufferUsageVertexBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
		r.vertexBuffer, r.vertexBufferMem)
	/*
		r.copyBuffer(stagingBuffer, r.vertexBuffer, bSize)

		vk.DestroyBuffer(*r.device, *stagingBuffer, nil)
		vk.FreeMemory(*r.device, *stagingBufferMem, nil)
		stagingBuffer = nil
		stagingBufferMem = nil*/

	r.updateVertex()
}

func (r *Renderer) copyBuffer(src *vk.Buffer, dst *vk.Buffer, size vk.DeviceSize) {
	commands := r.beginSingle()

	region := vk.BufferCopy{}

	region.SrcOffset = 0
	region.DstOffset = 0
	region.Size = size

	vk.CmdCopyBuffer(commands, *src, *dst, 1, []vk.BufferCopy{region})

	r.endSingle(commands)
}

func (r *Renderer) beginSingle() vk.CommandBuffer {
	allocInfo := vk.CommandBufferAllocateInfo{}

	allocInfo.SType = vk.StructureTypeCommandBufferAllocateInfo
	allocInfo.Level = vk.CommandBufferLevelPrimary
	allocInfo.CommandPool = *r.commandPool
	allocInfo.CommandBufferCount = 1

	commands := make([]vk.CommandBuffer, 1)

	vkr := vk.AllocateCommandBuffers(*r.device, &allocInfo, commands)

	if vkr != vk.Success {
		panic(vkr)
	}

	beginInfo := vk.CommandBufferBeginInfo{}

	beginInfo.SType = vk.StructureTypeCommandBufferBeginInfo
	beginInfo.Flags = vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit)

	vk.BeginCommandBuffer(commands[0], &beginInfo)

	return commands[0]
}

func (r *Renderer) endSingle(commands vk.CommandBuffer) {
	vk.EndCommandBuffer(commands)

	submitInfo := vk.SubmitInfo{}

	submitInfo.SType = vk.StructureTypeSubmitInfo
	submitInfo.CommandBufferCount = 1
	submitInfo.PCommandBuffers = []vk.CommandBuffer{commands}

	vk.QueueSubmit(*r.gfxq, 1, []vk.SubmitInfo{submitInfo}, vk.NullFence)
	vk.QueueWaitIdle(*r.gfxq)

	vk.FreeCommandBuffers(*r.device, *r.commandPool, 1, []vk.CommandBuffer{commands})
	commands = nil
}

func (r *Renderer) getAttributes() []vk.VertexInputAttributeDescription {
	descriptions := make([]vk.VertexInputAttributeDescription, 3)

	descriptions[0].Binding = 0 // pos
	descriptions[0].Location = 0
	descriptions[0].Format = vk.FormatR32g32b32a32Sfloat
	descriptions[0].Offset = 0

	descriptions[1].Binding = 0 // col
	descriptions[1].Location = 1
	descriptions[1].Format = vk.FormatR32g32b32a32Sfloat
	descriptions[1].Offset = vec4Size

	descriptions[2].Binding = 0 // tex
	descriptions[2].Location = 2
	descriptions[2].Format = vk.FormatR32g32b32a32Sfloat
	descriptions[2].Offset = vec4Size * 2

	return descriptions
}

func (r *Renderer) getBindings() []vk.VertexInputBindingDescription {
	binding := make([]vk.VertexInputBindingDescription, 1)

	binding[0].Binding = 0
	binding[0].Stride = vertSize * 4 // 8 floats (8 * 4 bytes)
	binding[0].InputRate = vk.VertexInputRateVertex

	return binding
}

func (r *Renderer) recreateSwap() {
	/*width, height := 0, 0
	for width == 0 || height == 0 {
		width, height = glfw.GetCurrentContext().GetSize()
		glfw.WaitEvents()
	}*/

	vk.DeviceWaitIdle(*r.device)

	r.cleanupSwap()

	r.createSwap()
	r.createViews()
	r.createRenderPass()
	r.createPipeline()
	r.createDepth()
	r.createFramebuffers()
	r.createUniformBuffers()
	r.createDescriptorPool()
	r.createDescriptorSets()
	r.createCommandBuffers()
}

func (r *Renderer) cleanupSwap() {
	vk.DestroyImageView(*r.device, *r.depthImageView, nil)
	vk.DestroyImage(*r.device, *r.depthImage, nil)
	vk.FreeMemory(*r.device, *r.depthImageMem, nil)

	for i := 0; i < len(r.framebuffers); i++ {
		vk.DestroyFramebuffer(*r.device, r.framebuffers[i], nil)
	}

	vk.FreeCommandBuffers(*r.device, *r.commandPool, uint32(len(r.commands)), r.commands)

	vk.DestroyPipeline(*r.device, r.pipeline, nil)
	vk.DestroyPipelineLayout(*r.device, *r.pipelineLayout, nil)
	vk.DestroyPipelineCache(*r.device, *r.pipelineCache, nil)

	vk.DestroyRenderPass(*r.device, *r.renderPass, nil)

	for i := 0; i < len(r.sciv); i++ {
		vk.DestroyImageView(*r.device, r.sciv[i], nil)
	}

	vk.DestroySwapchain(*r.device, *r.sc, nil)

	/*for i := 0; i < len(r.framebuffers); i++ {
		vk.DestroyFramebuffer(*r.device, r.framebuffers[i], nil)
	}*/

	vk.FreeCommandBuffers(*r.device, *r.commandPool, uint32(len(r.commands)), r.commands)

	//for i := 0; i < len(r.sci); i++ {
	vk.DestroyBuffer(*r.device, *r.uniformBuffers, nil)
	vk.FreeMemory(*r.device, *r.uniformBufferMem, nil)
	//}

	vk.DestroyDescriptorPool(*r.device, *r.descriptorPool, nil)
}

func (r *Renderer) createSync() {
	semInfo := vk.SemaphoreCreateInfo{}

	semInfo.SType = vk.StructureTypeSemaphoreCreateInfo

	fenInfo := vk.FenceCreateInfo{}

	fenInfo.SType = vk.StructureTypeFenceCreateInfo
	fenInfo.Flags = vk.FenceCreateFlags(vk.FenceCreateSignaledBit)

	r.available = make([]*vk.Semaphore, r.maxFrames)
	r.finished = make([]*vk.Semaphore, r.maxFrames)
	r.inf = make([]*vk.Fence, r.maxFrames)

	for i := 0; i < r.maxFrames; i++ {
		r.available[i] = new(vk.Semaphore)
		vkr := vk.CreateSemaphore(*r.device, &semInfo, nil, r.available[i])

		if vkr != vk.Success {
			panic(vkr)
		}

		r.finished[i] = new(vk.Semaphore)
		vkr = vk.CreateSemaphore(*r.device, &semInfo, nil, r.finished[i])

		if vkr != vk.Success {
			panic(vkr)
		}

		r.inf[i] = new(vk.Fence)
		vkr = vk.CreateFence(*r.device, &fenInfo, nil, r.inf[i])

		if vkr != vk.Success {
			panic(vkr)
		}
	}
}

func (r *Renderer) updateCommandBuffers() {
	//	vk.DeviceWaitIdle(*r.device)

	vk.ResetCommandPool(*r.device, *r.commandPool, vk.CommandPoolResetFlags(vk.CommandPoolResetReleaseResourcesBit))

	for i := 0; i < len(r.commands); i++ {
		//		vk.ResetCommandBuffer(r.commands[i], vk.CommandBufferResetFlags(vk.CommandBufferResetReleaseResourcesBit))

		info := vk.CommandBufferBeginInfo{}

		info.SType = vk.StructureTypeCommandBufferBeginInfo
		info.Flags = vk.CommandBufferUsageFlags(vk.CommandBufferUsageSimultaneousUseBit | vk.CommandBufferUsageOneTimeSubmitBit)
		//info.PInheritanceInfo = nil

		vkr := vk.BeginCommandBuffer(r.commands[i], &info)

		if vkr != vk.Success {
			panic(vkr)
		}

		pass := vk.RenderPassBeginInfo{}

		pass.SType = vk.StructureTypeRenderPassBeginInfo
		pass.RenderPass = *r.renderPass
		pass.Framebuffer = r.framebuffers[i]
		pass.RenderArea.Offset = vk.Offset2D{X: 0, Y: 0}
		pass.RenderArea.Extent = r.sce

		clear := make([]vk.ClearValue, 2)

		clear[0].SetColor([]float32{0.0, 0.03, 0.03, 1.0})
		clear[1].SetDepthStencil(1.0, 0)

		pass.ClearValueCount = uint32(len(clear))
		pass.PClearValues = clear

		vk.CmdBeginRenderPass(r.commands[i], &pass, vk.SubpassContentsInline) // start of render pass

		vk.CmdBindPipeline(r.commands[i], vk.PipelineBindPointGraphics, r.pipeline)

		vertexBuffers := []vk.Buffer{*r.vertexBuffer}
		offsets := []vk.DeviceSize{0}
		vk.CmdBindVertexBuffers(r.commands[i], 0, 1, vertexBuffers, offsets)

		vk.CmdBindIndexBuffer(r.commands[i], *r.indexBuffer, 0, vk.IndexTypeUint16)

		vk.CmdBindDescriptorSets(r.commands[i], vk.PipelineBindPointGraphics, *r.pipelineLayout, 0, 1, []vk.DescriptorSet{*r.descriptorSets}, 0, nil)

		//vk.CmdDrawIndexed(r.commands[i], uint32(len(r.indexData)), 1, 0, 0, 0) // draw command
		vk.CmdDrawIndexedIndirect(r.commands[i], *r.drawBuffer, 0, drawCount, uint32(unsafe.Sizeof(vk.DrawIndexedIndirectCommand{})))

		vk.CmdEndRenderPass(r.commands[i])

		vkr = vk.EndCommandBuffer(r.commands[i])

		if vkr != vk.Success {
			panic(vkr)
		}
	}
}

func (r *Renderer) createCommandBuffers() {
	r.commands = make([]vk.CommandBuffer, len(r.framebuffers))

	alloc := vk.CommandBufferAllocateInfo{}

	alloc.SType = vk.StructureTypeCommandBufferAllocateInfo
	alloc.CommandPool = *r.commandPool
	alloc.Level = vk.CommandBufferLevelPrimary
	alloc.CommandBufferCount = uint32(len(r.commands))

	vkr := vk.AllocateCommandBuffers(*r.device, &alloc, r.commands)

	if vkr != vk.Success {
		panic(vkr)
	}

	bSize := vk.DeviceSize(uint32(unsafe.Sizeof(vk.DrawIndexedIndirectCommand{})))

	r.drawBuffer = new(vk.Buffer)
	r.drawBufferMem = new(vk.DeviceMemory)

	r.createBuffer(bSize, vk.BufferUsageFlags(vk.BufferUsageIndirectBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
		r.drawBuffer, r.drawBufferMem)

	r.updateCommandBuffers()
}

func (r *Renderer) createCommandPool() {
	graphics, _ := r.findQueueFamilies(*r.physical)

	info := vk.CommandPoolCreateInfo{}
	info.SType = vk.StructureTypeCommandPoolCreateInfo
	info.QueueFamilyIndex = graphics
	info.Flags = vk.CommandPoolCreateFlags(vk.CommandPoolCreateTransientBit | vk.CommandPoolCreateResetCommandBufferBit)

	r.commandPool = new(vk.CommandPool)
	vkr := vk.CreateCommandPool(*r.device, &info, nil, r.commandPool)

	if vkr != vk.Success {
		panic(vkr)
	}
}

func (r *Renderer) createFramebuffers() {
	r.framebuffers = make([]vk.Framebuffer, len(r.sciv))

	for i := 0; i < len(r.sciv); i++ {
		attachments := []vk.ImageView{
			r.sciv[i],
			*r.depthImageView,
		}

		info := vk.FramebufferCreateInfo{}

		info.SType = vk.StructureTypeFramebufferCreateInfo
		info.RenderPass = *r.renderPass
		info.AttachmentCount = uint32(len(attachments))
		info.PAttachments = attachments
		info.Width = r.sce.Width
		info.Height = r.sce.Height
		info.Layers = 1

		vkr := vk.CreateFramebuffer(*r.device, &info, nil, &r.framebuffers[i])

		if vkr != vk.Success {
			panic(vkr)
		}
	}
}

func (r *Renderer) createRenderPass() {
	color := vk.AttachmentDescription{} // color buffer attachment

	color.Format = r.scif
	color.Samples = vk.SampleCount1Bit
	color.LoadOp = vk.AttachmentLoadOpClear // clear framebuffer
	color.StoreOp = vk.AttachmentStoreOpStore
	color.StencilLoadOp = vk.AttachmentLoadOpDontCare
	color.StencilStoreOp = vk.AttachmentStoreOpDontCare
	color.InitialLayout = vk.ImageLayoutUndefined
	color.FinalLayout = vk.ImageLayoutPresentSrc // present in swap chain

	colorRef := vk.AttachmentReference{} // reference to color

	colorRef.Attachment = 0 // layout in shader
	colorRef.Layout = vk.ImageLayoutColorAttachmentOptimal

	depth := vk.AttachmentDescription{}
	depth.Format = vk.FormatD32Sfloat
	depth.Samples = vk.SampleCount1Bit
	depth.LoadOp = vk.AttachmentLoadOpClear
	depth.StoreOp = vk.AttachmentStoreOpDontCare
	depth.StencilLoadOp = vk.AttachmentLoadOpDontCare
	depth.StencilStoreOp = vk.AttachmentStoreOpDontCare
	depth.InitialLayout = vk.ImageLayoutUndefined
	depth.FinalLayout = vk.ImageLayoutDepthStencilAttachmentOptimal

	depthRef := vk.AttachmentReference{}
	depthRef.Attachment = 1
	depthRef.Layout = vk.ImageLayoutDepthStencilAttachmentOptimal

	subpass := vk.SubpassDescription{}

	subpass.PipelineBindPoint = vk.PipelineBindPointGraphics
	subpass.ColorAttachmentCount = 1
	subpass.PColorAttachments = []vk.AttachmentReference{colorRef}
	subpass.PDepthStencilAttachment = &depthRef

	dep := vk.SubpassDependency{}

	dep.SrcSubpass = vk.SubpassExternal
	dep.DstSubpass = 0
	dep.SrcStageMask = vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)
	dep.SrcAccessMask = 0
	dep.DstStageMask = vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)
	dep.DstAccessMask = vk.AccessFlags(vk.AccessColorAttachmentReadBit | vk.AccessColorAttachmentWriteBit)

	attachments := []vk.AttachmentDescription{color, depth}

	info := vk.RenderPassCreateInfo{}

	info.SType = vk.StructureTypeRenderPassCreateInfo
	info.AttachmentCount = uint32(len(attachments))
	info.PAttachments = attachments
	info.SubpassCount = 1
	info.PSubpasses = []vk.SubpassDescription{subpass}
	info.DependencyCount = 1
	info.PDependencies = []vk.SubpassDependency{dep}

	r.renderPass = new(vk.RenderPass)
	vkr := vk.CreateRenderPass(*r.device, &info, nil, r.renderPass)

	if vkr != vk.Success {
		panic(vkr)
	}
}

func (r *Renderer) createPipeline() {
	var vkr vk.Result

	vertSrc := readFile("shader/vert.spv")
	fragSrc := readFile("shader/frag.spv")

	readFile("shader/main.frag")

	vertMod := r.createShaderModule(vertSrc)
	fragMod := r.createShaderModule(fragSrc)

	vertInfo := vk.PipelineShaderStageCreateInfo{}
	vertInfo.SType = vk.StructureTypePipelineShaderStageCreateInfo
	vertInfo.Stage = vk.ShaderStageVertexBit
	vertInfo.Module = vertMod
	vertInfo.PName = "main\x00" // entrypoint

	fragInfo := vk.PipelineShaderStageCreateInfo{}
	fragInfo.SType = vk.StructureTypePipelineShaderStageCreateInfo
	fragInfo.Stage = vk.ShaderStageFragmentBit
	fragInfo.Module = fragMod
	fragInfo.PName = "main\x00" // entrypoint

	stages := []vk.PipelineShaderStageCreateInfo{vertInfo, fragInfo}

	vertInputInfo := vk.PipelineVertexInputStateCreateInfo{}
	vertInputInfo.SType = vk.StructureTypePipelineVertexInputStateCreateInfo
	vertInputInfo.VertexBindingDescriptionCount = 1
	vertInputInfo.PVertexBindingDescriptions = r.getBindings()

	attributes := r.getAttributes()

	vertInputInfo.VertexAttributeDescriptionCount = uint32(len(attributes))
	vertInputInfo.PVertexAttributeDescriptions = attributes

	inputAsm := vk.PipelineInputAssemblyStateCreateInfo{}

	inputAsm.SType = vk.StructureTypePipelineInputAssemblyStateCreateInfo
	inputAsm.Topology = vk.PrimitiveTopologyTriangleList
	inputAsm.PrimitiveRestartEnable = vk.False

	viewport := vk.Viewport{}
	viewport.X = 0.0
	viewport.Y = 0.0
	viewport.Width = float32(r.sce.Width)
	viewport.Height = float32(r.sce.Height)
	viewport.MinDepth = 0.0
	viewport.MaxDepth = 1.0

	scissor := vk.Rect2D{}

	scissor.Offset = vk.Offset2D{
		X: 0,
		Y: 0,
	}
	scissor.Extent = r.sce

	viewportState := vk.PipelineViewportStateCreateInfo{}

	viewportState.SType = vk.StructureTypePipelineViewportStateCreateInfo
	viewportState.ViewportCount = 1
	viewportState.PViewports = []vk.Viewport{viewport}
	viewportState.ScissorCount = 1
	viewportState.PScissors = []vk.Rect2D{scissor}

	rasterizer := vk.PipelineRasterizationStateCreateInfo{}

	rasterizer.SType = vk.StructureTypePipelineRasterizationStateCreateInfo
	rasterizer.DepthClampEnable = vk.False // clamp values instead of discard
	rasterizer.RasterizerDiscardEnable = vk.False
	rasterizer.PolygonMode = vk.PolygonModeFill
	rasterizer.LineWidth = 1.0
	//rasterizer.CullMode = vk.CullModeFlags(vk.CullModeBackBit) // culling
	rasterizer.CullMode = vk.CullModeFlags(vk.CullModeNone) // culling
	rasterizer.FrontFace = vk.FrontFaceClockwise               // winding
	rasterizer.DepthBiasEnable = vk.False                      // depth biasing
	rasterizer.DepthBiasConstantFactor = 0.0
	rasterizer.DepthBiasClamp = 0.0
	rasterizer.DepthBiasSlopeFactor = 0.0

	multisampling := vk.PipelineMultisampleStateCreateInfo{}

	multisampling.SType = vk.StructureTypePipelineMultisampleStateCreateInfo
	multisampling.SampleShadingEnable = vk.False
	multisampling.RasterizationSamples = vk.SampleCount1Bit
	multisampling.MinSampleShading = 1.0
	multisampling.PSampleMask = nil
	multisampling.AlphaToCoverageEnable = vk.False
	multisampling.AlphaToOneEnable = vk.False

	depth := vk.PipelineDepthStencilStateCreateInfo{}
	depth.SType = vk.StructureTypePipelineDepthStencilStateCreateInfo
	depth.DepthTestEnable = vk.True
	depth.DepthWriteEnable = vk.True

	depth.DepthCompareOp = vk.CompareOpLess
	depth.DepthBoundsTestEnable = vk.False
	depth.MinDepthBounds = 0.0
	depth.MaxDepthBounds = 1.0
	depth.StencilTestEnable = vk.False

	blend := vk.PipelineColorBlendAttachmentState{}

	blend.ColorWriteMask = vk.ColorComponentFlags(
		vk.ColorComponentRBit |
			vk.ColorComponentGBit |
			vk.ColorComponentBBit |
			vk.ColorComponentABit)

	blend.BlendEnable = vk.False
	blend.SrcColorBlendFactor = vk.BlendFactorOne
	blend.DstColorBlendFactor = vk.BlendFactorZero
	blend.ColorBlendOp = vk.BlendOpAdd
	blend.SrcAlphaBlendFactor = vk.BlendFactorOne
	blend.DstAlphaBlendFactor = vk.BlendFactorZero
	blend.AlphaBlendOp = vk.BlendOpAdd

	blending := vk.PipelineColorBlendStateCreateInfo{}

	blending.SType = vk.StructureTypePipelineColorBlendStateCreateInfo
	blending.LogicOpEnable = vk.False
	blending.LogicOp = vk.LogicOpCopy
	blending.AttachmentCount = 1
	blending.PAttachments = []vk.PipelineColorBlendAttachmentState{blend}
	blending.BlendConstants[0] = 0.0
	blending.BlendConstants[1] = 0.0
	blending.BlendConstants[2] = 0.0
	blending.BlendConstants[3] = 0.0

	dynStates := []vk.DynamicState{vk.DynamicStateViewport, vk.DynamicStateLineWidth}

	dynState := vk.PipelineDynamicStateCreateInfo{} // viewport, line width and (blend) can be changed using this

	dynState.SType = vk.StructureTypePipelineDynamicStateCreateInfo
	dynState.DynamicStateCount = 2
	dynState.PDynamicStates = dynStates

	layoutInfo := vk.PipelineLayoutCreateInfo{}

	layoutInfo.SType = vk.StructureTypePipelineLayoutCreateInfo
	layoutInfo.SetLayoutCount = 1
	layoutInfo.PSetLayouts = []vk.DescriptorSetLayout{*r.descriptorLayout}
	layoutInfo.PushConstantRangeCount = 0
	layoutInfo.PPushConstantRanges = nil

	r.pipelineLayout = new(vk.PipelineLayout)
	vkr = vk.CreatePipelineLayout(*r.device, &layoutInfo, nil, r.pipelineLayout)

	if vkr != vk.Success {
		panic(vkr)
	}

	r.pipelineCache = new(vk.PipelineCache)
	vkr = vk.CreatePipelineCache(*r.device, &vk.PipelineCacheCreateInfo{
		SType: vk.StructureTypePipelineCacheCreateInfo,
	}, nil, r.pipelineCache)

	info := vk.GraphicsPipelineCreateInfo{}

	info.SType = vk.StructureTypeGraphicsPipelineCreateInfo
	info.StageCount = 2
	info.PStages = stages

	info.PVertexInputState = &vertInputInfo
	info.PInputAssemblyState = &inputAsm
	info.PViewportState = &viewportState
	info.PRasterizationState = &rasterizer
	info.PMultisampleState = &multisampling
	info.PDepthStencilState = &depth
	info.PColorBlendState = &blending
	info.PDynamicState = nil
	info.PTessellationState = nil

	info.Layout = *r.pipelineLayout
	info.RenderPass = *r.renderPass
	info.Subpass = 0
	//info.BasePipelineHandle = nil
	//info.BasePipelineIndex = -1

	pipelines := make([]vk.Pipeline, 1)
	//infos := make([]vk.GraphicsPipelineCreateInfo, 1)
	//infos[0] = info
	vkr = vk.CreateGraphicsPipelines(*r.device, vk.NullPipelineCache, 1,
		[]vk.GraphicsPipelineCreateInfo{info}, nil, pipelines)
	r.pipeline = pipelines[0]

	if vkr != vk.Success {
		panic(vkr)
	}

	vk.DestroyShaderModule(*r.device, vertMod, nil)
	vk.DestroyShaderModule(*r.device, fragMod, nil)

	println("graphics pipeline: ", r.pipeline)
}

func (r *Renderer) createShaderModule(src []byte) vk.ShaderModule {
	info := vk.ShaderModuleCreateInfo{}

	info.SType = vk.StructureTypeShaderModuleCreateInfo
	/*var s []uint32

	for i := 0; i < len(src); i++ {
		s = append(s, uint32(src[i]))
	}*/
	//info.CodeSize = uint(len(s))
	//info.PCode = s

	info.CodeSize = uint(len(src))
	info.PCode = sliceUint32(src)

	var module vk.ShaderModule

	vkr := vk.CreateShaderModule(*r.device, &info, nil, &module)

	if vkr != vk.Success {
		panic(vkr)
	}

	return module
}

func (r *Renderer) createViews() {
	r.sciv = make([]vk.ImageView, len(r.sci))

	for i := 0; i < len(r.sci); i++ {
		info := vk.ImageViewCreateInfo{}

		info.SType = vk.StructureTypeImageViewCreateInfo
		info.Image = r.sci[i]
		info.ViewType = vk.ImageViewType2d
		info.Format = r.scif

		info.Components.R = vk.ComponentSwizzleIdentity
		info.Components.G = vk.ComponentSwizzleIdentity
		info.Components.B = vk.ComponentSwizzleIdentity
		info.Components.A = vk.ComponentSwizzleIdentity

		info.SubresourceRange.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
		info.SubresourceRange.BaseMipLevel = 0
		info.SubresourceRange.LevelCount = 1
		info.SubresourceRange.BaseArrayLayer = 0
		info.SubresourceRange.LayerCount = 1

		vkr := vk.CreateImageView(*r.device, &info, nil, &r.sciv[i])

		if vkr != vk.Success {
			panic(vkr)
		}
	}
}

func (r *Renderer) createSwap() {
	var capabilities vk.SurfaceCapabilities

	vk.GetPhysicalDeviceSurfaceCapabilities(*r.physical, r.surface, &capabilities)
	capabilities.Deref()

	var formatCount uint32

	vk.GetPhysicalDeviceSurfaceFormats(*r.physical, r.surface, &formatCount, nil)

	formats := make([]vk.SurfaceFormat, formatCount)

	vk.GetPhysicalDeviceSurfaceFormats(*r.physical, r.surface, &formatCount, formats) // assuming count != 0

	var presentCount uint32

	vk.GetPhysicalDeviceSurfacePresentModes(*r.physical, r.surface, &presentCount, nil)

	presents := make([]vk.PresentMode, presentCount)

	vk.GetPhysicalDeviceSurfacePresentModes(*r.physical, r.surface, &presentCount, presents) // assuming again

	capabilities.CurrentExtent.Deref()
	extent := capabilities.CurrentExtent // assuming typical

	imageCount := capabilities.MinImageCount + 1

	var info vk.SwapchainCreateInfo

	formats[0].Deref()

	info.SType = vk.StructureTypeSwapchainCreateInfo
	info.Surface = r.surface
	info.MinImageCount = imageCount
	info.ImageFormat = formats[0].Format // just pick first?
	info.ImageColorSpace = formats[0].ColorSpace
	info.ImageExtent = extent
	info.ImageArrayLayers = 1
	info.ImageUsage = vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit)

	info.ImageSharingMode = vk.SharingModeExclusive // if graphicsfamily == presentfamily

	info.PreTransform = capabilities.CurrentTransform
	info.CompositeAlpha = vk.CompositeAlphaOpaqueBit
	//info.PresentMode = presents[0]
	//info.PresentMode = vk.PresentModeFifo // vsync
	info.PresentMode = vk.PresentModeImmediate // swap interval 0
	info.Clipped = vk.True
	info.OldSwapchain = vk.NullSwapchain

	r.sc = new(vk.Swapchain)
	vkr := vk.CreateSwapchain(*r.device, &info, nil, r.sc)

	if vkr != vk.Success {
		panic(vkr)
	}

	var images uint32
	vk.GetSwapchainImages(*r.device, *r.sc, &images, nil)

	r.sci = make([]vk.Image, images)

	vk.GetSwapchainImages(*r.device, *r.sc, &images, r.sci) // swap chain image handles

	r.scif = formats[0].Format
	r.sce = extent

	println("swap chain: ", r.sc)
}

func validationLayers() (names []string) {
	var count uint32
	vk.EnumerateInstanceLayerProperties(&count, nil)
	list := make([]vk.LayerProperties, count)
	vk.EnumerateInstanceLayerProperties(&count, list)
	for _, layer := range list {
		layer.Deref()
		names = append(names, vk.ToString(layer.LayerName[:]))
	}
	return names
}

// INSTANCE CREATION
func (r *Renderer) createInstance() {
	appinfo := vk.ApplicationInfo{}
	appinfo.SType = vk.StructureTypeApplicationInfo
	appinfo.PApplicationName = "GRM"
	appinfo.ApplicationVersion = vk.MakeVersion(1, 0, 0)
	appinfo.PEngineName = "GRM"
	appinfo.EngineVersion = vk.MakeVersion(1, 0, 0)
	appinfo.ApiVersion = vk.ApiVersion10

	info := vk.InstanceCreateInfo{}

	info.SType = vk.StructureTypeInstanceCreateInfo
	info.PApplicationInfo = &appinfo

	info.EnabledExtensionCount = (uint32)(len(r.extensions))
	info.PpEnabledExtensionNames = r.extensions

	info.EnabledLayerCount = uint32(len(r.validation)) // validation layers here
	info.PpEnabledLayerNames = r.validation

	//var vki vk.Instance
	r.vki = new(vk.Instance)

	vkr := vk.CreateInstance(&info, nil, r.vki)

	if vkr != vk.Success {
		panic(vkr)
	}

	vk.InitInstance(*r.vki)

	//r.vki = &vki

	println("created instance: ", *r.vki)
}

// PHYSICAL DEVICE
func (r *Renderer) getPhysical() {
	var deviceCount uint32

	//var devices []vk.PhysicalDevice

	vk.EnumeratePhysicalDevices(*r.vki, &deviceCount, nil)

	println("devices: ", deviceCount)

	if deviceCount == 0 {
		panic(deviceCount)
	}

	devices := make([]vk.PhysicalDevice, deviceCount)

	vkr := vk.EnumeratePhysicalDevices(*r.vki, &deviceCount, devices)

	println("d: ", devices, vkr)

	r.physical = new(vk.PhysicalDevice)
	*r.physical = nil

	for _, e := range devices {
		if r.deviceSuitable(e) {
			*r.physical = e
		}
	}

	if *r.physical == nil {
		*r.physical = devices[0] // device selection failed
	}
	*r.physical = devices[0] // just pick 0 for now

	println("selected device: ", *r.physical)
}

func (r *Renderer) findQueues() uint32 {
	var count uint32

	vk.GetPhysicalDeviceQueueFamilyProperties(*r.physical, &count, nil)

	properties := make([]vk.QueueFamilyProperties, count)

	vk.GetPhysicalDeviceQueueFamilyProperties(*r.physical, &count, properties)

	var indices uint32

	for i, e := range properties {
		println("qc: ", e.QueueCount)
		if e.QueueCount > 0 && (int(e.QueueFlags)&int(vk.QueueGraphicsBit) == 1) { // only in C ref
			indices = uint32(i)
		}
	}

	return indices
}

func (r *Renderer) getExtensions() []string {
	//var found int

	var count uint32

	var ext []string

	vk.EnumerateDeviceExtensionProperties(*r.physical, "", &count, nil)

	available := make([]vk.ExtensionProperties, count)

	vk.EnumerateDeviceExtensionProperties(*r.physical, "", &count, available)

	//found = 0
	//	required := make([]string, len(r.dextensions))
	for i := uint32(0); i < count; i++ {
		available[i].Deref()

		/*for j := 0; j < len(r.dextensions); j++ {
			var req [256]byte

			for k := 0; k < 256; k++ {
				req[k] = 0
				if k < len(r.dextensions[j]) {
					req[k] = r.dextensions[j][k]
				}
			}

			//println(string(available[i].ExtensionName[:]))
			if available[i].ExtensionName == req {
				found++
			}
		}*/
		found := false
		for j := uint32(0); j < uint32(len(r.dextensions)); j++ {
			if unsafeEqual(r.dextensions[j], available[i].ExtensionName[:]) {
				found = true
			}
		}
		_ = found
		//if found {
		ext = append(ext, string(available[i].ExtensionName[:]))
		//}
	}

	return ext
}

func (r *Renderer) findQueueFamilies(device vk.PhysicalDevice) (uint32, uint32) {
	var familyCount uint32
	vk.GetPhysicalDeviceQueueFamilyProperties(*r.physical, &familyCount, nil)

	families := make([]vk.QueueFamilyProperties, familyCount)

	vk.GetPhysicalDeviceQueueFamilyProperties(*r.physical, &familyCount, families)

	var graphics uint32
	var present uint32

	for i := uint32(0); i < familyCount; i++ {
		families[i].Deref()

		var req vk.QueueFlags
		req |= vk.QueueFlags(vk.QueueGraphicsBit)

		res := families[i].QueueFlags & req

		if families[i].QueueCount > 0 && res != 0 {
			graphics = i
		}

		var presentSupport vk.Bool32
		presentSupport = vk.False

		vk.GetPhysicalDeviceSurfaceSupport(*r.physical, i, r.surface, &presentSupport)

		if families[i].QueueCount > 0 && presentSupport.B() {
			present = i
		}
	}

	return graphics, present
}

func (r *Renderer) createLogical() {
	graphics, present := r.findQueueFamilies(*r.physical)

	println("graphics i ", graphics)
	println("present i ", present)

	vkqi := []vk.DeviceQueueCreateInfo{{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueFamilyIndex: graphics,
		QueueCount:       1,
		PQueuePriorities: []float32{1.0},
	}}
	vkqi = append(vkqi, vk.DeviceQueueCreateInfo{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueFamilyIndex: present,
		QueueCount:       1,
		PQueuePriorities: []float32{1.0},
	})

	features := make([]vk.PhysicalDeviceFeatures, 1)

	var info vk.DeviceCreateInfo
	info.SType = vk.StructureTypeDeviceCreateInfo

	info.QueueCreateInfoCount = (uint32)(len(vkqi))
	info.PQueueCreateInfos = vkqi
	info.PEnabledFeatures = features
	//info.EnabledExtensionCount = (uint32)(len(r.dextensions))
	//info.PpEnabledExtensionNames = r.dextensions
	info.EnabledExtensionCount = uint32(1)
	info.PpEnabledExtensionNames = []string{string([]byte(vk.KhrSwapchainExtensionName))}
	info.EnabledLayerCount = (uint32)(len(r.validation))
	info.PpEnabledLayerNames = r.validation

	r.device = new(vk.Device)
	vkr := vk.CreateDevice(*r.physical, &info, nil, r.device)

	if vkr != vk.Success {
		panic(vkr)
	}

	println("logical device: ", *r.device)

	var graphicsQueue vk.Queue
	vk.GetDeviceQueue(*r.device, graphics, 0, &graphicsQueue)
	r.gfxq = &graphicsQueue

	var presentQueue vk.Queue
	vk.GetDeviceQueue(*r.device, present, 0, &presentQueue)
	r.pq = &presentQueue
}

func (r *Renderer) deviceSuitable(device vk.PhysicalDevice) bool {
	var returns bool
	var properties vk.PhysicalDeviceProperties
	var features vk.PhysicalDeviceFeatures

	vk.GetPhysicalDeviceProperties(device, &properties)
	properties.Deref() // deref to read c values
	vk.GetPhysicalDeviceFeatures(device, &features)
	features.Deref()

	println(string(properties.DeviceName[:]))

	returns = properties.DeviceType == vk.PhysicalDeviceTypeDiscreteGpu //&& (features.GeometryShader == 1) // bool32

	println("ds: ", device, returns)

	return returns
}

func (r *Renderer) getTriCount() uint32 {
	return uint32(len(r.vertexData)) / vertSize
}

func (r *Renderer) getVertexSize() int {
	return int(unsafe.Sizeof(r.vertexData[0])) * len(r.vertexData)
}

func (r *Renderer) getUBOSize() int {
	return int(unsafe.Sizeof(UniformBufferObject{}))
}

func (r *Renderer) updateUniform(cur uint32) {
	uData := uboData(r.ubo)

	var data unsafe.Pointer
	vk.MapMemory(*r.device, *r.uniformBufferMem, 0, vk.DeviceSize(len(uData)), 0, &data)
	vk.Memcopy(data, uData)
	vk.UnmapMemory(*r.device, *r.uniformBufferMem)
}

func (r *Renderer) destroy() { // finish this later
	r.cleanupSwap()

	vk.DestroyDescriptorSetLayout(*r.device, *r.descriptorLayout, nil)

	vk.DestroyBuffer(*r.device, *r.vertexBuffer, nil)
	//vk.DestroyBuffer(*r.device, *r.colBuffer, nil)
	//vk.DestroyBuffer(*r.device, *r.texBuffer, nil)
	vk.DestroyBuffer(*r.device, *r.indexBuffer, nil)

	vk.FreeMemory(*r.device, *r.vertexBufferMem, nil)
	vk.FreeMemory(*r.device, *r.indexBufferMem, nil)

	vk.DestroySurface(*r.vki, r.surface, nil)
	vk.DestroyDevice(*r.device, nil)
	vk.DestroyInstance(*r.vki, nil)

	for i := 0; i < len(r.framebuffers); i++ {
		vk.DestroyFramebuffer(*r.device, r.framebuffers[i], nil)
	}

	vk.DestroyCommandPool(*r.device, *r.commandPool, nil)

	for i := 0; i < r.maxFrames; i++ {
		vk.DestroySemaphore(*r.device, *r.available[i], nil)
		vk.DestroySemaphore(*r.device, *r.finished[i], nil)
	}
}

func (r *Renderer) findMemoryType(filter uint32, properties vk.MemoryPropertyFlags) uint32 {
	memProp := vk.PhysicalDeviceMemoryProperties{}
	vk.GetPhysicalDeviceMemoryProperties(*r.physical, &memProp)

	memProp.Deref()

	for i := uint32(0); i < memProp.MemoryTypeCount; i++ {
		memProp.MemoryTypes[i].Deref()
		if /*filter & (1 << i) == 1 && */ (memProp.MemoryTypes[i].PropertyFlags & properties) == properties {
			return i
		}
	}

	return 0
}

func (r *Renderer) createBuffer(size vk.DeviceSize, usage vk.BufferUsageFlags, properties vk.MemoryPropertyFlags, buffer *vk.Buffer, bufferMem *vk.DeviceMemory) {
	info := vk.BufferCreateInfo{}

	info.SType = vk.StructureTypeBufferCreateInfo
	info.Size = size
	info.Usage = usage
	info.SharingMode = vk.SharingModeExclusive

	vkr := vk.CreateBuffer(*r.device, &info, nil, buffer)

	if vkr != vk.Success {
		panic(vkr)
	}

	memReq := vk.MemoryRequirements{}

	vk.GetBufferMemoryRequirements(*r.device, *buffer, &memReq)
	memReq.Deref()

	allocInfo := vk.MemoryAllocateInfo{}

	allocInfo.SType = vk.StructureTypeMemoryAllocateInfo
	allocInfo.AllocationSize = memReq.Size
	allocInfo.MemoryTypeIndex = r.findMemoryType(memReq.MemoryTypeBits, properties)

	vkr = vk.AllocateMemory(*r.device, &allocInfo, nil, bufferMem)

	if vkr != vk.Success {
		panic(vkr)
	}

	vk.BindBufferMemory(*r.device, *buffer, *bufferMem, 0)
}

func unsafeEqual(a string, b []byte) bool {
	bbp := *(*string)(unsafe.Pointer(&b))
	return a == bbp
}
