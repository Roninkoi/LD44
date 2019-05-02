package game

import (
	"../grm"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"strconv"
)

type GUI struct {
	textSprite grm.Sprite
	reaper     grm.Sprite
	boss       grm.Sprite

	logo grm.Sprite

	scythe      grm.Sprite
	scytheTicks float64

	soulMeter grm.Sprite
	soulBar   grm.Sprite
	soulNum   int
	soulPer   float64

	money    grm.Sprite
	moneyNum int

	quotaNum  int
	billNum   int
	salaryNum int
	shiftNum  int

	day int

	clockFace  grm.Sprite
	clockCover grm.Sprite

	ticks float64

	clockTicks float64
}

func (g *GUI) renderText(r *grm.Renderer, s string, x float32, y float32, scale float32) { // sprite 158, 0
	xOffs := (float32)(0.0)
	for i := 0; i < len(s); i++ {
		xOffs += 1.0 * (((6.0 / 11.0) * scale) / 0.55)
		ai := []rune(string(s[i]))[0]
		x0 := (ai - 32) % 16
		y0 := (int)(math.Floor((float64)(ai-32) / 16.0))

		g.textSprite.AnimLoad([]int{0}, 1.0, []mgl32.Vec4{
			{(float32)(x0*6 + 158.0), (float32)(y0 * 11), 6.0, 11.0}})

		g.textSprite.Mesh.Model = mgl32.Translate3D(xOffs+x, y, 0.12)
		g.textSprite.Mesh.Model = g.textSprite.Mesh.Model.Mul4(mgl32.Scale3D(((6.0/11.0)*scale)/0.55, scale, 1.0))
		g.textSprite.Mesh.Model = g.textSprite.Mesh.Model.Mul4(mgl32.HomogRotate3DY(float32(1.57)))
		g.textSprite.Mesh.DisableTransform()
		g.textSprite.Mesh.Update()

		g.textSprite.AnimDraw(r)
	}
}

func (g *GUI) load() {
	g.textSprite.LoadTextSprite(nil, "")

	g.reaper.LoadSprite(nil, "")
	g.reaper.AnimLoad([]int{0}, 1.0, []mgl32.Vec4{{11.0, 187.0, 205.0, 226.0}})
	g.reaper.Mesh.Model = g.reaper.Mesh.Model.Mul4(mgl32.HomogRotate3DY(1.57))
	g.reaper.Mesh.Model = g.reaper.Mesh.Model.Mul4(mgl32.Translate3D(0.1, 0.0, 0.0))
	g.reaper.Mesh.Update()

	g.boss.LoadSprite(nil, "")
	g.boss.AnimLoad([]int{0}, 1.0, []mgl32.Vec4{{920.0, 183.0, 289.0, 198.0}})
	g.boss.Mesh.Model = g.boss.Mesh.Model.Mul4(mgl32.Translate3D(0.0, -0.4, 0.13))
	g.boss.Mesh.Model = g.boss.Mesh.Model.Mul4(mgl32.HomogRotate3DY(1.57))
	g.boss.Mesh.Model = g.boss.Mesh.Model.Mul4(mgl32.Scale3D(1.0*0.75, 1.45*0.75, 1.0))
	g.boss.Mesh.DisableTransform()
	g.boss.Mesh.Update()

	g.logo.LoadSprite(nil, "")
	g.logo.AnimLoad([]int{0}, 1.0, []mgl32.Vec4{{1659.0, 581.0, 329.0, 329.0}})
	g.logo.Mesh.Model = g.logo.Mesh.Model.Mul4(mgl32.Translate3D(0.05, -0.1, 0.2))
	g.logo.Mesh.Model = g.logo.Mesh.Model.Mul4(mgl32.HomogRotate3DY(1.5707))
	g.logo.Mesh.Model = g.logo.Mesh.Model.Mul4(mgl32.Scale3D(1.0, 1.77, 1.0))
	g.logo.Mesh.DisableTransform()
	g.logo.Mesh.Update()

	g.scythe.LoadSprite(nil, "")
	g.scythe.AnimLoad([]int{0}, 1.0, []mgl32.Vec4{{307.0, 0.0, 160.0, 160.0}})
	g.scythe.Mesh.Model = g.reaper.Mesh.Model.Mul4(mgl32.Translate3D(0.0, 0.0, -0.5))
	g.scythe.Mesh.DisableTransform()
	g.scythe.Mesh.Update()

	g.soulMeter.LoadSprite(nil, "")
	g.soulMeter.AnimLoad([]int{0}, 1.0, []mgl32.Vec4{{307.0, 154.0, 42.0, 271.0}})
	g.soulMeter.Mesh.Model = g.soulMeter.Mesh.Model.Mul4(mgl32.Scale3D(0.23, 0.23*6.45, 1.0))
	g.soulMeter.Mesh.Model = g.soulMeter.Mesh.Model.Mul4(mgl32.Translate3D(3.85, 0.0, 0.05))
	g.soulMeter.Mesh.Model = g.soulMeter.Mesh.Model.Mul4(mgl32.HomogRotate3DY(float32(1.57)))
	g.soulMeter.Mesh.DisableTransform()
	g.soulMeter.Mesh.Update()

	g.soulBar.LoadSprite(nil, "")
	g.soulBar.AnimLoad([]int{0}, 1.0, []mgl32.Vec4{{361.0, 195.0, 21.0, 220.0}})
	g.soulBar.Mesh.Model = g.soulBar.Mesh.Model.Mul4(mgl32.Scale3D(0.23, 0.23*10.47, 1.0))
	g.soulBar.Mesh.Model = g.soulBar.Mesh.Model.Mul4(mgl32.Translate3D(3.85, 0.0, 0.1))
	g.soulBar.Mesh.Model = g.soulBar.Mesh.Model.Mul4(mgl32.HomogRotate3DY(float32(1.57)))
	g.soulBar.Mesh.DisableTransform()
	g.soulBar.Mesh.Update()

	g.money.LoadSprite(nil, "")
	g.money.AnimLoad([]int{0}, 1.0, []mgl32.Vec4{{432.0, 175.0, 50.0, 50.0}})
	g.money.Mesh.Model = g.money.Mesh.Model.Mul4(mgl32.Scale3D(0.2, 0.2, 1.0))
	g.money.Mesh.Model = g.money.Mesh.Model.Mul4(mgl32.Translate3D(-4.2, -2.0, 0.1))
	g.money.Mesh.Model = g.money.Mesh.Model.Mul4(mgl32.HomogRotate3DY(float32(1.57)))
	g.money.Mesh.DisableTransform()
	g.money.Mesh.Update()

	g.clockFace.LoadSprite(nil, "")
	g.clockFace.AnimLoad([]int{0}, 1.0, []mgl32.Vec4{{425.0, 339.0, 72.0, 72.0}})
	g.clockFace.Mesh.Model = g.clockFace.Mesh.Model.Mul4(mgl32.Scale3D(0.6, 0.6, 1.0))
	g.clockFace.Mesh.Model = g.clockFace.Mesh.Model.Mul4(mgl32.Translate3D(-1.0, -1.6, 0.1))
	g.clockFace.Mesh.Model = g.clockFace.Mesh.Model.Mul4(mgl32.HomogRotate3DY(float32(1.57)))
	g.clockFace.Mesh.DisableTransform()
	g.clockFace.Mesh.Update()

	g.clockCover.LoadSprite(nil, "")
	g.clockCover.AnimLoad([]int{0}, 1.0, []mgl32.Vec4{{529.0, 339.0, 72.0, 72.0}})
	g.clockCover.Mesh.Model = g.clockCover.Mesh.Model.Mul4(mgl32.Scale3D(0.6, 0.6, 1.0))
	g.clockCover.Mesh.Model = g.clockCover.Mesh.Model.Mul4(mgl32.Translate3D(-1.05, -1.6, 0.05))
	g.clockCover.Mesh.Model = g.clockCover.Mesh.Model.Mul4(mgl32.HomogRotate3DY(float32(1.57)))
	g.clockCover.Mesh.DisableTransform()
	g.clockCover.Mesh.Update()
}

func (g *GUI) tick() {
	g.scytheTicks += 0.1
}

func (g *GUI) introScreen(r *grm.Renderer) {
	g.logo.AnimDraw(r)

	if int(g.ticks/30.0)%2 == 0 {
		g.renderText(r, "Press SPACE to start", -0.7, 0.85, 0.07)
	}
}

func (g *GUI) statScreen(r *grm.Renderer) {
	g.boss.AnimDraw(r)

	g.renderText(r, "Your quota is: "+strconv.Itoa(g.quotaNum)+" souls", -0.6, 0.3-0.1, 0.05)

	g.renderText(r, "Your salary is: "+strconv.Itoa(g.salaryNum), -0.50, 0.45-0.1, 0.05)

	g.renderText(r, "The last bill was: "+strconv.Itoa(g.billNum), -0.55, 0.6-0.1, 0.05)

	g.renderText(r, "The next shift will be "+strconv.Itoa(g.shiftNum)+" hours", -0.75, 0.75-0.1, 0.05)

	g.renderText(r, "SPACE to start next day", -0.55, 0.9, 0.05)
}

func (g *GUI) startScreen(r *grm.Renderer) {
	g.statScreen(r)

	g.renderText(r, "Greetings new employee.", -0.55, 0.0, 0.05)
}

func (g *GUI) winScreen(r *grm.Renderer) {
	g.renderText(r, "Great job.", -0.25, 0.0, 0.05)

	g.statScreen(r)
}

func (g *GUI) loseScreen(r *grm.Renderer) {
	g.boss.AnimDraw(r)

	g.renderText(r, "You're fired.", -0.35, 0.3, 0.05)

	g.renderText(r, "SPACE to restart", -0.43, 0.55, 0.05)
}

func (g *GUI) draw(r *grm.Renderer) {
	g.scythe.Mesh.Model = mgl32.Ident4()
	g.scythe.Mesh.Model = g.scythe.Mesh.Model.Mul4(mgl32.Scale3D(1.2, 1.2, 1.0))
	g.scythe.Mesh.Model = g.scythe.Mesh.Model.Mul4(mgl32.Translate3D(-1.05, 0.72, 0.5))
	g.scythe.Mesh.Model = g.scythe.Mesh.Model.Mul4(mgl32.HomogRotate3DY(float32(1.57)))

	if g.scytheTicks > 6.28 {
		g.scytheTicks = 6.28
	}

	arc := float32(math.Sin(g.scytheTicks)) * 0.8
	g.scythe.Mesh.Model = g.scythe.Mesh.Model.Mul4(mgl32.HomogRotate3DX(0.0 + arc*0.8))
	g.scythe.Mesh.Model = g.scythe.Mesh.Model.Mul4(mgl32.Translate3D(0.0, -0.3, 0.6-1.6*arc))
	g.scythe.Mesh.Update()

	g.scythe.AnimDraw(r)

	g.soulMeter.AnimDraw(r)

	g.soulBar.Mesh.Model = mgl32.Ident4()
	scale := float32(g.soulPer)
	if scale > 1.0 {
		scale = 1.0
	}
	g.soulBar.Mesh.Model = g.soulBar.Mesh.Model.Mul4(mgl32.Scale3D(0.125, 0.125*9.8*scale, 1.0))
	g.soulBar.Mesh.Model = g.soulBar.Mesh.Model.Mul4(mgl32.Translate3D(7.0, 0.07+0.47*0.125*9.8*(1.0-scale)/scale, 0.1))
	g.soulBar.Mesh.Model = g.soulBar.Mesh.Model.Mul4(mgl32.HomogRotate3DY(float32(1.57)))
	g.soulBar.Mesh.Update()

	g.soulBar.AnimDraw(r)

	g.renderText(r, strconv.Itoa(int(math.Floor(g.soulPer*100.0)))+"%", 0.75, 0.85, 0.05)

	g.money.AnimDraw(r)
	g.renderText(r, strconv.Itoa(g.moneyNum), -0.74, -0.42, 0.05)
	g.renderText(r, "Day: "+strconv.Itoa(g.day), -0.94, -0.2, 0.05)

	g.clockFace.Mesh.Model = mgl32.Ident4()
	g.clockFace.Mesh.Model = g.clockFace.Mesh.Model.Mul4(mgl32.Scale3D(0.57, 0.57, 1.0))
	g.clockFace.Mesh.Model = g.clockFace.Mesh.Model.Mul4(mgl32.Translate3D(-1.05*(0.6/0.57)-0.007, -1.6*(0.6/0.57)-0.007, 0.1))
	g.clockFace.Mesh.Model = g.clockFace.Mesh.Model.Mul4(mgl32.HomogRotate3DY(float32(1.57)))
	g.clockFace.Mesh.Model = g.clockFace.Mesh.Model.Mul4(mgl32.HomogRotate3DX(float32(g.clockTicks)))
	g.clockFace.Mesh.Update()
	g.clockFace.AnimDraw(r)

	g.clockCover.AnimDraw(r)
}
