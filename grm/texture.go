package grm

import (
	vk "github.com/vulkan-go/vulkan"
	"unsafe"
)

type Texture struct {
	image    *vk.Image
	imageMem *vk.DeviceMemory
	imageView *vk.ImageView
	sampler *vk.Sampler

	Path string

	W uint32
	H uint32
}

func (t *Texture) createSampler(r *Renderer) {
	info := vk.SamplerCreateInfo{}
	info.SType = vk.StructureTypeSamplerCreateInfo
	info.MagFilter = vk.FilterNearest // PIXEL ART!
	info.MinFilter = vk.FilterNearest

	info.AddressModeU = vk.SamplerAddressModeClampToBorder // black outside
	info.AddressModeV = vk.SamplerAddressModeClampToBorder
	info.AddressModeW = vk.SamplerAddressModeClampToBorder

	info.AnisotropyEnable = vk.False
	info.MaxAnisotropy = 0

	info.BorderColor = vk.BorderColorIntOpaqueBlack
	info.UnnormalizedCoordinates = vk.False

	info.CompareEnable = vk.False
	info.CompareOp = vk.CompareOpAlways

	info.MipmapMode = vk.SamplerMipmapModeNearest
	info.MipLodBias = 0.0
	info.MinLod = 0.0
	info.MaxLod = 0.0

	t.sampler = new(vk.Sampler)
	vkr := vk.CreateSampler(*r.device, &info, nil, t.sampler)

	if vkr != vk.Success {
		panic(vkr)
	}
}

func (t *Texture) createImageView(r *Renderer) {
	info := vk.ImageViewCreateInfo{}
	info.SType = vk.StructureTypeImageViewCreateInfo
	info.Image = *t.image
	info.ViewType = vk.ImageViewType2d
	info.Format = vk.FormatR8g8b8a8Unorm
	info.SubresourceRange.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
	info.SubresourceRange.BaseMipLevel = 0
	info.SubresourceRange.LevelCount = 1
	info.SubresourceRange.BaseArrayLayer = 0
	info.SubresourceRange.LayerCount = 1

	t.imageView = new(vk.ImageView)
	vkr := vk.CreateImageView(*r.device, &info, nil, t.imageView)

	if vkr != vk.Success {
		panic(vkr)
	}
}

func (t *Texture) createTextureImage(r *Renderer) {
	var dat []byte
	var err error
	dat, t.W, t.H, err = loadTextureData(t.Path, 0)
	dat, t.W, t.H, err = loadTextureData(t.Path, int(t.W*4))

	if err != nil {
		panic(err)
	}
	iSize := vk.DeviceSize(len(dat))

	var stagingBuffer *vk.Buffer
	var stagingBufferMem *vk.DeviceMemory
	stagingBuffer = new(vk.Buffer)
	stagingBufferMem = new(vk.DeviceMemory)
	r.createBuffer(iSize, vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit),
		stagingBuffer, stagingBufferMem)

	var data unsafe.Pointer
	vk.MapMemory(*r.device, *stagingBufferMem, 0, iSize, 0, &data)
	vk.Memcopy(data, dat)
	vk.UnmapMemory(*r.device, *stagingBufferMem)

	t.image = new(vk.Image)
	t.imageMem = new(vk.DeviceMemory)
	r.createImage(t.W, t.H, vk.FormatR8g8b8a8Unorm, vk.ImageTilingOptimal,
		vk.ImageUsageFlags(vk.ImageUsageTransferDstBit|vk.ImageUsageSampledBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyDeviceLocalBit),
		t.image, t.imageMem)

	r.transitionLayout(*t.image, vk.FormatR8g8b8a8Unorm, vk.ImageLayoutUndefined, vk.ImageLayoutTransferDstOptimal)

	t.copyBufferImage(r, *stagingBuffer, *t.image, t.W, t.H)

	r.transitionLayout(*t.image, vk.FormatR8g8b8a8Unorm, vk.ImageLayoutTransferDstOptimal, vk.ImageLayoutShaderReadOnlyOptimal)

	vk.DestroyBuffer(*r.device, *stagingBuffer, nil)
	vk.FreeMemory(*r.device, *stagingBufferMem, nil)
	stagingBuffer = nil
	stagingBufferMem = nil
}

func (t *Texture) copyBufferImage(r *Renderer, buffer vk.Buffer, image vk.Image, w uint32, h uint32) {
	commands := r.beginSingle()

	region := vk.BufferImageCopy{}
	region.BufferOffset = 0
	region.BufferRowLength = 0
	region.BufferImageHeight = 0

	region.ImageSubresource.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
	region.ImageSubresource.MipLevel = 0
	region.ImageSubresource.BaseArrayLayer = 0
	region.ImageSubresource.LayerCount = 1

	offs := vk.Offset3D{}

	offs.X = 0
	offs.Y = 0
	offs.Z = 0

	ext := vk.Extent3D{}

	ext.Width = w
	ext.Height = h
	ext.Depth = 1

	region.ImageOffset = offs
	region.ImageExtent = ext

	vk.CmdCopyBufferToImage(commands, buffer, image,
		vk.ImageLayoutTransferDstOptimal, 1, []vk.BufferImageCopy{region})

	r.endSingle(commands)
}

func (t *Texture) Load(r *Renderer, path string) {
	t.Path = path

	t.createTextureImage(r)
	t.createImageView(r)
	t.createSampler(r)
}

func (t *Texture) Destroy(r *Renderer) {
	vk.DestroySampler(*r.device, *t.sampler, nil)
	vk.DestroyImageView(*r.device, *t.imageView, nil)
	vk.DestroyImage(*r.device, *t.image, nil)
	vk.FreeMemory(*r.device, *t.imageMem, nil)
}

func (t *Texture) copyMem(r *Renderer, another *Texture) {
	var dat unsafe.Pointer
	var data1 unsafe.Pointer
	size := (another.W) * (another.H) * 4

	vk.MapMemory(*r.device, *another.imageMem, 0, vk.DeviceSize(size), 0, &data1)
	vk.Memcopy(dat, unsafeByte(data1, int(size)))
	vk.UnmapMemory(*r.device, *another.imageMem)

	var data2 unsafe.Pointer
	vk.MapMemory(*r.device, *t.imageMem, 0, vk.DeviceSize(size), 0, &data2)
	vk.Memcopy(data2, unsafeByte(dat, int(size)))
	vk.UnmapMemory(*r.device, *t.imageMem)
}

func (r *Renderer) createImage(w uint32, h uint32, format vk.Format, tiling vk.ImageTiling, usage vk.ImageUsageFlags, props vk.MemoryPropertyFlags, image *vk.Image, imageMem *vk.DeviceMemory) {
	info := vk.ImageCreateInfo{}
	info.SType = vk.StructureTypeImageCreateInfo
	info.ImageType = vk.ImageType2d
	info.Extent.Width = w
	info.Extent.Height = h
	info.Extent.Depth = 1
	info.MipLevels = 1
	info.ArrayLayers = 1

	info.Format = format
	info.Tiling = tiling // VK_IMAGE_TILING_LINEAR
	info.InitialLayout = vk.ImageLayoutUndefined
	info.Usage = usage
	info.SharingMode = vk.SharingModeExclusive

	info.Samples = vk.SampleCount1Bit
	info.Flags = 0

	vkr := vk.CreateImage(*r.device, &info, nil, image)

	if vkr != vk.Success {
		panic(vkr)
	}

	req := vk.MemoryRequirements{}
	vk.GetImageMemoryRequirements(*r.device, *image, &req)
	req.Deref()

	allocInfo := vk.MemoryAllocateInfo{}
	allocInfo.SType = vk.StructureTypeMemoryAllocateInfo
	allocInfo.AllocationSize = req.Size
	allocInfo.MemoryTypeIndex = r.findMemoryType(req.MemoryTypeBits, props)

	vkr = vk.AllocateMemory(*r.device, &allocInfo, nil, imageMem)

	vk.BindImageMemory(*r.device, *image, *imageMem, 0)
}

func (r *Renderer) transitionLayout(image vk.Image, format vk.Format, old vk.ImageLayout, new vk.ImageLayout) {
	commands := r.beginSingle()

	barrier := vk.ImageMemoryBarrier{}
	barrier.SType = vk.StructureTypeImageMemoryBarrier
	barrier.OldLayout = old
	barrier.NewLayout = new

	barrier.SrcQueueFamilyIndex = vk.QueueFamilyIgnored
	barrier.DstQueueFamilyIndex = vk.QueueFamilyIgnored

	barrier.Image = image
	barrier.SubresourceRange.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
	barrier.SubresourceRange.BaseMipLevel = 0
	barrier.SubresourceRange.LevelCount = 1
	barrier.SubresourceRange.BaseArrayLayer = 0
	barrier.SubresourceRange.LayerCount = 1

	barrier.SrcAccessMask = 0
	barrier.DstAccessMask = 0

	if new == vk.ImageLayoutDepthStencilAttachmentOptimal {
		barrier.SubresourceRange.AspectMask = vk.ImageAspectFlags(vk.ImageAspectDepthBit)
	} else {
		barrier.SubresourceRange.AspectMask = vk.ImageAspectFlags(vk.ImageAspectColorBit)
	}

	var srcStage vk.PipelineStageFlags
	var dstStage vk.PipelineStageFlags

	if old == vk.ImageLayoutUndefined && new == vk.ImageLayoutTransferDstOptimal {
		barrier.SrcAccessMask = 0
		barrier.DstAccessMask = vk.AccessFlags(vk.AccessTransferWriteBit)

		srcStage = vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit)
		dstStage = vk.PipelineStageFlags(vk.PipelineStageTransferBit)
	} else if old == vk.ImageLayoutTransferDstOptimal && new == vk.ImageLayoutShaderReadOnlyOptimal {
		barrier.SrcAccessMask = vk.AccessFlags(vk.AccessTransferWriteBit)
		barrier.DstAccessMask = vk.AccessFlags(vk.AccessShaderReadBit)

		srcStage = vk.PipelineStageFlags(vk.PipelineStageTransferBit)
		dstStage = vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit)
	} else if old == vk.ImageLayoutUndefined && new == vk.ImageLayoutDepthStencilAttachmentOptimal {
		barrier.SrcAccessMask = 0
		barrier.DstAccessMask = vk.AccessFlags(vk.AccessDepthStencilAttachmentReadBit | vk.AccessDepthStencilAttachmentWriteBit)

		srcStage = vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit)
		dstStage = vk.PipelineStageFlags(vk.PipelineStageEarlyFragmentTestsBit)
	} else {
		panic("bad layout transition!")
	}

	vk.CmdPipelineBarrier(commands,
		srcStage, dstStage,
		0,
		0, nil,
		0, nil,
		1, []vk.ImageMemoryBarrier{barrier})

	r.endSingle(commands)
}
