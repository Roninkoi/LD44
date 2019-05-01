package grm

import (
	"github.com/vulkan-go/glfw/v3.3/glfw"
	vk "github.com/vulkan-go/vulkan"
	"time"
)

type Gorium struct {
	Running bool

	Window Window

	Renderer Renderer

	Input Input

	Ticks    float64
	fps      int
	fpsTicks int
	fpsTime  float64
	tickTime float64

	rt float64
	tt float64

	time float64

	run func()
	draw func()
	quit func()
}

// start execution (exported)
func (g *Gorium) Start(run func(), draw func(), quit func()) {
	println("Gorium (GRM) v0.1")

	g.run = run
	g.draw = draw
	g.quit = quit

	g.init()
	g.main()
}

// initialize and load engine
func (g *Gorium) init() {
	g.Running = true

	g.Window.init()
	g.Window.create()

	vk.SetGetInstanceProcAddr(glfw.GetVulkanGetInstanceProcAddress())

	g.Renderer.init(g.Window.Window)
}

func timeNow() float64 {
	return (float64)(time.Now().UnixNano()) / 1000000.0
}

// main engine program
func (g *Gorium) main() {
	for g.Running {
		g.time = timeNow()

		if g.time-g.fpsTime >= 1000.0 {
			g.fpsTime = g.time

			g.fps = g.fpsTicks

			println("fps: ", g.fps, " rt: ", g.rt, " tt: ", g.tt, "dc: ", g.Renderer.drawCount)

			g.fpsTicks = 0
		}

		if g.time-g.tickTime >= 16.6 {
			g.tickTime = g.time

			g.tick()
			g.render()
		}

		g.Running = !g.Window.closed()

		g.Window.update()

		//g.render()
	}

	g.cleanup()

	g.quit()
}

func (g *Gorium) render() {
	g.fpsTicks++
	g.rt = timeNow()

	g.Renderer.ticks = g.Ticks

	g.draw() // game draw

	g.Renderer.Draw()

	g.rt = timeNow() - g.rt
}

func (g *Gorium) tick() {
	g.tt = timeNow()

	g.Ticks += 1.0

	g.Input.getKeys(g.Window.Window)
	g.run() // game logic

	g.tt = timeNow() - g.tt
}

// clean up
func (g *Gorium) cleanup() {
	vk.DeviceWaitIdle(*g.Renderer.device)

	//g.renderer.destroy()
}

// destroy engine (exported)
func (g *Gorium) Destroy() {
	g.Window.destroy()

	glfw.Terminate()
}
