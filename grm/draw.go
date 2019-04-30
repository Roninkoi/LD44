package grm

import (
	"github.com/go-gl/mathgl/mgl32"
	vk "github.com/vulkan-go/vulkan"
)

func (r *Renderer) Clear() {
	r.vertexData = nil
	r.indexData = nil
	r.drawCount = r.draws
	r.draws = 0
}

func (r *Renderer) AddData(vertices *[]float32, indices *[]uint16) {
	r.draws++
	r.vertexData = append(r.vertexData, *vertices...)
	r.indexData = append(r.indexData, *indices...)
}

func (r *Renderer) Add(vertices *[]float32, texes *[]float32, colors *[]float32, indices *[]uint16, tex *Texture) {
	if len(r.indexData)+len(*indices) >= batchSize || len(r.vertexData)+len(*vertices) >= batchSize {
		r.Flush()
	}
	if tex != nil {
		if tex.Path != r.texture.Path {
			//r.Flush()
		}
		//r.texture = tex
	}

	leaved := make([]float32, len(*vertices)*3)

	if texes != nil {
		for i := 0; i < len(*vertices)/4; i += 1 {
			leaved[i*12] = (*vertices)[i*4]
			leaved[i*12+1] = (*vertices)[i*4+1]
			leaved[i*12+2] = (*vertices)[i*4+2]
			leaved[i*12+3] = (*vertices)[i*4+3]

			if colors != nil && len(*colors) == len(*vertices) {
				leaved[i*12+4] = (*colors)[i*4+0]
				leaved[i*12+5] = (*colors)[i*4+1]
				leaved[i*12+6] = (*colors)[i*4+2]
				leaved[i*12+7] = (*colors)[i*4+3]
			} else {
				leaved[i*12+4] = 1.0
				leaved[i*12+5] = 1.0
				leaved[i*12+6] = 1.0
				leaved[i*12+7] = 1.0
			}

			leaved[i*12+8] = (*texes)[i*4]
			leaved[i*12+9] = 1.0 - (*texes)[i*4+1]
			leaved[i*12+10] = (*texes)[i*4+2]
			leaved[i*12+11] = (*texes)[i*4+3]
		}
	} else { // assume that vertex data contains texes
		leaved = *vertices
	}
	iSize := len(r.indexData)
	vSize := len(r.vertexData) / vertSize

	//r.vertexData = append(r.vertexData, leaved...)
	//r.indexData = append(r.indexData, *indices...)
	r.AddData(&leaved, indices)

	for i := iSize; i < len(r.indexData); i++ {
		r.indexData[i] += uint16(vSize)
	}
}

func (r *Renderer) SetCam(pos mgl32.Vec3, rot mgl32.Vec3) {
	r.ubo.cam = mgl32.Ident4()

	rotM := mgl32.Ident4()
	rotM = rotM.Mul4(mgl32.HomogRotate3DX(rot.X()))
	rotM = rotM.Mul4(mgl32.HomogRotate3DY(rot.Y()))
	rotM = rotM.Mul4(mgl32.HomogRotate3DZ(rot.Z()))

	r.ubo.cam = r.ubo.cam.Mul4(rotM)

	posM := mgl32.Ident4()
	posM = posM.Mul4(mgl32.Translate3D(-pos.X(), -pos.Y(), -pos.Z()))

	r.ubo.cam = r.ubo.cam.Mul4(posM)
}

func (r *Renderer) Flush() {
	r.updateData()

	r.Draw()
	r.Clear()
	//r.updateTex()
}

func (r *Renderer) updateData() {
	r.updateVertex()
	r.updateIndex()
	r.updateDraw()
}

func (r *Renderer) Update(cur uint32) {
	r.ubo.obj = mgl32.Ident4()

	//r.ubo.cam = mgl32.LookAt(float32(math.Sin(r.ticks/100.0))*5.0, 0.0, float32(math.Cos(r.ticks/100.0))*5.0, 0.0, float32(math.Cos(r.ticks/200.0))*5.0, 0.0, 0.0, -1.0, 0.0)
	//r.ubo.cam = mgl32.LookAt(0.0, 4.0, 0.0, 0.0, 0.0, 1.57, 0.0, -1.0, 0.0)

	r.ubo.proj = mgl32.Perspective(1.57, float32(r.sce.Width/r.sce.Height), 0.1, 100.0)

	r.ubo.amb = r.Ambient

	r.updateUniform(cur)

	r.updateData()

	r.Clear()

	//r.updateCommandBuffers()
}

func (r *Renderer) Draw() {
	vk.WaitForFences(*r.device, 1, []vk.Fence{*r.inf[r.frame]}, vk.True, vk.MaxUint64)

	vk.ResetFences(*r.device, 1, []vk.Fence{*r.inf[r.frame]})

	var index uint32

	result := vk.AcquireNextImage(*r.device, *r.sc, vk.MaxUint64, *r.available[r.frame], vk.NullFence, &index)

	if result == vk.ErrorOutOfDate {
		r.recreateSwap()
	}

	r.Update(index)

	submit := vk.SubmitInfo{}
	submit.SType = vk.StructureTypeSubmitInfo

	waitSem := []vk.Semaphore{*r.available[r.frame]}
	waitSta := []vk.PipelineStageFlags{vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)}

	submit.WaitSemaphoreCount = 1
	submit.PWaitSemaphores = waitSem
	submit.PWaitDstStageMask = waitSta

	submit.CommandBufferCount = 1
	submit.PCommandBuffers = []vk.CommandBuffer{r.commands[index]}

	signal := []vk.Semaphore{*r.finished[r.frame]}

	submit.SignalSemaphoreCount = 1
	submit.PSignalSemaphores = signal

	var vkr vk.Result

	vkr = vk.QueueSubmit(*r.gfxq, 1, []vk.SubmitInfo{submit}, *r.inf[r.frame])

	if vkr != vk.Success {
		panic(vkr)
	}

	present := vk.PresentInfo{}

	present.SType = vk.StructureTypePresentInfo
	present.WaitSemaphoreCount = 1
	present.PWaitSemaphores = signal

	present.SwapchainCount = 1
	present.PSwapchains = []vk.Swapchain{*r.sc}
	present.PImageIndices = []uint32{index}
	present.PResults = nil

	result = vk.QueuePresent(*r.pq, &present)

	if result == vk.ErrorOutOfDate || result == vk.Suboptimal { // do explicit resize later
		r.recreateSwap()
	} else if result != vk.Success {
		panic(vkr)
	}

	r.frame = (r.frame + 1) % r.maxFrames
}
