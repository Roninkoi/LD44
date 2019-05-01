package grm

import "github.com/vulkan-go/glfw/v3.3/glfw"

type Input struct {
	W bool
	A bool
	S bool
	D bool

	Up bool
	Down bool
	Left bool
	Right bool

	F bool
	R bool

	E bool
	EPress bool
	Q bool

	Space bool
	SpacePress bool
}

func (i *Input) getKeys(win *glfw.Window) {
	if win.GetKey(glfw.KeyW) == glfw.Press || win.GetKey(glfw.KeyZ) == glfw.Press {
		i.W = true
	} else {
		i.W = false
	}
	if win.GetKey(glfw.KeyA) == glfw.Press || win.GetKey(glfw.KeyQ) == glfw.Press {
		i.A = true
	} else {
		i.A = false
	}
	if win.GetKey(glfw.KeyS) == glfw.Press {
		i.S = true
	} else {
		i.S = false
	}
	if win.GetKey(glfw.KeyD) == glfw.Press {
		i.D = true
	} else {
		i.D = false
	}

	if win.GetKey(glfw.KeyUp) == glfw.Press {
		i.Up = true
	} else {
		i.Up = false
	}
	if win.GetKey(glfw.KeyDown) == glfw.Press {
		i.Down = true
	} else {
		i.Down = false
	}
	if win.GetKey(glfw.KeyLeft) == glfw.Press {
		i.Left = true
	} else {
		i.Left = false
	}
	if win.GetKey(glfw.KeyRight) == glfw.Press {
		i.Right = true
	} else {
		i.Right = false
	}

	if win.GetKey(glfw.KeyF) == glfw.Press {
		i.F = true
	} else {
		i.F = false
	}
	if win.GetKey(glfw.KeyR) == glfw.Press {
		i.R = true
	} else {
		i.R = false
	}

	if win.GetKey(glfw.KeyE) == glfw.Press {
		if !i.E {
			i.EPress = true
		} else {
			i.EPress = false
		}
		i.E = true
	} else {
		i.E = false
		i.EPress = false
	}
	if win.GetKey(glfw.KeyQ) == glfw.Press {
		i.Q = true
	} else {
		i.Q = false
	}

	if win.GetKey(glfw.KeySpace) == glfw.Press {
		if !i.Space {
			i.SpacePress = true
		} else {
			i.SpacePress = false
		}
		i.Space = true
	} else {
		i.Space = false
		i.SpacePress = false
	}
}
