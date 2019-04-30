package grm

import (
	"github.com/vulkan-go/glfw/v3.3/glfw"
)

type Window struct {
	Window *glfw.Window

	W int
	H int
}

func (w *Window) init() {
	err := glfw.Init()

	if err != nil {
		panic(err)
	}

	w.W = 1600
	w.H = 900
}

func (w *Window) create() {
	var err error

	glfw.WindowHint(glfw.ClientAPI, glfw.NoAPI)
	w.Window, err = glfw.CreateWindow(w.W, w.H, "GRM", nil, nil)

	if err != nil {
			panic(err)
	}

	//w.window.MakeContextCurrent() // for gl
}

func (w *Window) destroy() {
	w.Window.Destroy()
}

func (w *Window) closed() bool {
	return w.Window.ShouldClose()
}

func (w *Window) update() {
	if !w.Window.ShouldClose() {
		//w.window.SwapBuffers()
		glfw.PollEvents()
	}
}
