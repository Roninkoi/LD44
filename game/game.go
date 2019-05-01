package game

import (
	"Gorium/grm"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"math/rand"
)

type Game struct {
	e       grm.Gorium // ENGINE
	running bool

	world World

	player *Entity

	gui GUI

	win  bool
	lose bool

	gameEnded bool
	intro     bool

	debug bool
}

func (g *Game) load() {
	g.world.player.load()
	g.player = &g.world.player
	g.gui.load()

	g.gameEnded = true
	g.win = false
	g.lose = false
	g.intro = true

	g.debug = false // DEBUG

	g.world.load(&g.e.Renderer)
}

func (g *Game) run() { // GAME LOGIC
	if !g.running {
		g.load()
		g.running = true
	}

	if int(g.e.Ticks)%360 == 0 {
		println("pos ", g.player.pos.X(), g.player.pos.Y(), g.player.pos.Z())
		println("time ", g.world.time, "money ", g.world.moneyNum, "souls ", g.world.soulsNum, "quota ", g.world.quotaNum)
	}

	g.input()

	g.world.time += 1.0 / 1000.0
	g.e.Renderer.Ambient = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}.Mul(float32(math.Max(0.0, -0.3*math.Cos(g.world.time))) + 0.1)
	duskr := float32(math.Abs(0.3*math.Sin(g.world.time))) * float32(math.Abs(0.3*math.Sin(g.world.time))) * 2.0 * 1.5
	duskg := float32(math.Abs(0.3*math.Sin(g.world.time))) * float32(math.Abs(0.3*math.Sin(g.world.time))) * 1.5
	duskr += float32(math.Abs(0.3*math.Sin(g.world.time+1.57))) * 0.3
	duskg += float32(math.Abs(0.3*math.Sin(g.world.time+1.57))) * 0.3
	nightb := float32(math.Max(0.0, 0.3*math.Cos(g.world.time)))
	g.e.Renderer.Ambient = g.e.Renderer.Ambient.Add(mgl32.Vec4{
		duskr, duskg, nightb, 1.0})
	g.e.Renderer.Ambient = g.e.Renderer.Ambient.Mul(2.0)

	g.gui.clockTicks = g.world.time

	g.gui.moneyNum = g.world.moneyNum
	g.gui.soulPer = float64(g.world.soulsNum) / float64(g.world.quotaNum)
	g.gui.soulNum = g.world.soulsNum
	g.gui.quotaNum = g.world.quotaNum
	g.gui.billNum = g.world.billNum
	g.gui.shiftNum = g.world.shiftNum
	g.gui.salaryNum = g.world.salaryNum
	g.gui.day = g.world.day

	g.world.tick()

	g.gui.tick()

	if g.world.time > (float64(g.world.shiftNum)/24.0)*9.42 && !g.gameEnded && !g.debug { // GAME OVER
		g.gameEnded = true

		g.world.moneyNum -= g.world.billNum
		g.world.moneyNum += g.world.salaryNum
		g.world.shiftNum = int(((rand.Float64() + 1.0) / 2.0) * 24.0)

		g.world.salaryNum += int(math.Round(rand.Float64()))
		g.world.quotaNum += int((math.Round(rand.Float64() - 0.4)) * 3.0) // ???

		g.world.billNum += 1
		g.world.day += 1

		if g.world.soulsNum >= g.world.quotaNum && g.world.moneyNum > 0 {
			g.win = true
		} else {
			g.lose = true
		}
	}
}

func (g *Game) input() {
	if !g.gameEnded {
		moved := false
		movV := grm.Nv3()
		if g.e.Input.W {
			movV[0] += g.player.spd * float32(math.Sin(float64(g.player.rot[1])))
			movV[2] += -g.player.spd * float32(math.Cos(float64(g.player.rot[1])))
			moved = true
		}
		if g.e.Input.S {
			movV[0] += -g.player.spd * float32(math.Sin(float64(g.player.rot[1])))
			movV[2] += g.player.spd * float32(math.Cos(float64(g.player.rot[1])))
			moved = true
		}
		if g.e.Input.A {
			movV[0] += -g.player.spd * float32(math.Cos(float64(g.player.rot[1])))
			movV[2] += -g.player.spd * float32(math.Sin(float64(g.player.rot[1])))
			moved = true
		}
		if g.e.Input.D {
			movV[0] += g.player.spd * float32(math.Cos(float64(g.player.rot[1])))
			movV[2] += g.player.spd * float32(math.Sin(float64(g.player.rot[1])))
			moved = true
		}
		if g.e.Input.F && g.debug {
			movV[1] += g.player.spd
			moved = true
		}
		if g.e.Input.R && g.debug {
			movV[1] += -g.player.spd
			moved = true
		}

		//println(g.player.obj.IDist)
		if moved {
			g.player.obj.Phys.V = g.player.obj.Phys.V.Add(movV)
		}

		if g.e.Input.SpacePress {
			//g.player.obj.Phys.Pos[1] += -0.2
			//g.player.obj.Phys.V[1] += -0.2
			//g.world.test()

			g.gui.scytheTicks = 0.0
		}
		g.player.attacking = g.gui.scytheTicks < 4.0

		g.player.obj.Phys.V[0] *= 0.95
		g.player.obj.Phys.V[2] *= 0.95

		if g.e.Input.Up && g.debug {
			g.player.rot[0] += 0.05
		}
		if g.e.Input.Down && g.debug {
			g.player.rot[0] -= 0.05
		}
		if g.e.Input.Left {
			g.player.rot[1] -= 0.05
		}
		if g.e.Input.Right {
			g.player.rot[1] += 0.05
		}
	} else if g.win {
		if g.e.Input.SpacePress {
			g.gameEnded = false
			g.win = false
			g.lose = false
			g.world.start() // keep old stuff
		}
	} else if g.lose {
		if g.e.Input.SpacePress {
			g.gameEnded = false
			g.win = false
			g.lose = false
			g.world.restart()
		}
	} else if g.intro {
		if g.e.Input.SpacePress {
			g.intro = false
			g.win = false
			g.lose = false
		}
	} else {
		if g.e.Input.SpacePress {
			g.gameEnded = false
			g.win = false
			g.lose = false
			g.world.restart()
		}
	}
}

func (g *Game) draw() { // GAME DRAW
	g.e.Renderer.SetCam(g.player.obj.Phys.RPos, *g.player.rot)

	g.gui.ticks = g.e.Ticks
	if !g.gameEnded {
		g.gui.draw(&g.e.Renderer)
	} else if g.win {
		g.gui.winScreen(&g.e.Renderer)
	} else if g.lose {
		g.gui.loseScreen(&g.e.Renderer)
	} else if g.intro {
		g.gui.introScreen(&g.e.Renderer)
	} else {
		g.gui.startScreen(&g.e.Renderer)
	}

	g.world.draw(&g.e.Renderer)
}

func (g *Game) quit() {
	g.e.Destroy()
}

func (g *Game) Start() {
	g.e = grm.Gorium{}
	g.running = false

	g.e.Start(g.run, g.draw, g.quit)
}
