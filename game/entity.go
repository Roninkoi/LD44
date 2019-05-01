package game

import (
	"Gorium/grm"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"math/rand"
)

type Entity struct {
	obj grm.Obj

	sprite grm.Sprite

	pos *mgl32.Vec3
	rot *mgl32.Vec3

	spd float32

	hasAnim bool

	attacking bool

	movDir mgl32.Vec3
	movTicks int

	evil bool
}

func (e *Entity) load() {
	e.obj.Init()
	e.spd = 0.03
	e.pos = &e.obj.Phys.Pos
	e.rot = &e.obj.Phys.Rot
	e.hasAnim = false
	e.obj.Phys.RenderPos = true
	e.obj.Phys.V[1] = -0.07

	e.obj.SphereIsect = true
	e.obj.HasHull = false
	e.obj.Phys.IsStatic = false
}

func (e *Entity) draw(r *grm.Renderer) {
	e.sprite.AnimDraw(r)
}

func (e *Entity) tick() {
	if e.hasAnim {
		e.sprite.AnimUpdate()
	}

	e.movTicks++
	if e.movTicks > 60 {
		e.movTicks = 0
		e.movDir = mgl32.Vec3{e.spd*float32(rand.Float64()-0.5), 0.0, e.spd*0.2*float32(rand.Float64()-0.5)}
	}

	e.obj.Phys.V = e.obj.Phys.V.Add(e.movDir)
	e.obj.Phys.V[0] *= 0.9
	e.obj.Phys.V[2] *= 0.9

	e.obj.Mesh.SphereRadius = 1.0
	e.obj.Mesh.SphereCenter = e.obj.Phys.Pos

	e.sprite.Mesh.Model = mgl32.Ident4()
	e.sprite.Mesh.Model = e.sprite.Mesh.Model.Mul4(mgl32.Translate3D(float32(e.obj.Phys.Pos[0]), float32(e.obj.Phys.Pos[1]-0.5), float32(e.obj.Phys.Pos[2])))
	e.sprite.Mesh.Model = e.sprite.Mesh.Model.Mul4(mgl32.Scale3D(2.0, 2.75*2.0, 1.0))
	e.sprite.Mesh.Model = e.sprite.Mesh.Model.Mul4(mgl32.HomogRotate3DY(e.obj.Phys.Rot[1]))
	e.sprite.Mesh.Update()
}

func (e *Entity) randomChar(x0 float64, y0 float64, z0 float64) {
	n := math.Round(rand.Float64() * 2.0)
	e.hasAnim = true

	var c grm.Sprite
	c.LoadSprite(nil, "")

	if n == 0 {
		c.AnimLoad([]int{0, 1, 0, 2}, 250.0, []mgl32.Vec4{
			{692.0, 591.0, 16.0, 44.0},
			{692.0, 646.0, 16.0, 44.0},
			{692.0, 701.0, 16.0, 44.0},

			{782.0, 591.0, 16.0, 44.0},
			{782.0, 646.0, 16.0, 44.0},
			{782.0, 701.0, 16.0, 44.0},
		})
	} else if n == 1 {
		c.AnimLoad([]int{0, 1, 0, 2}, 250.0, []mgl32.Vec4{
			{721.0, 591.0, 16.0, 44.0},
			{721.0, 646.0, 16.0, 44.0},
			{721.0, 701.0, 16.0, 44.0},

			{808.0, 591.0, 16.0, 44.0},
			{808.0, 646.0, 16.0, 44.0},
			{808.0, 701.0, 16.0, 44.0},
		})
	} else {
		c.LoadSprite(nil, "")
		c.AnimLoad([]int{0, 1, 0, 2}, 250.0, []mgl32.Vec4{
			{752.0, 591.0, 16.0, 44.0},
			{752.0, 646.0, 16.0, 44.0},
			{752.0, 701.0, 16.0, 44.0},

			{838.0, 591.0, 16.0, 44.0},
			{838.0, 646.0, 16.0, 44.0},
			{838.0, 701.0, 16.0, 44.0},
		})
	}

	c.Mesh.Model = c.Mesh.Model.Mul4(mgl32.Scale3D(2.0, 2.75*2.0, 1.0))
	c.Mesh.Update()

	e.obj.Phys.Pos = mgl32.Vec3{float32(x0), float32(y0), float32(z0)}

	e.evil = false
	if rand.Float64() > 0.5 {
		e.evil = true
		c.Mesh.SetCol(mgl32.Vec4{1.5, 1.0, 1.0, 1.0})
	}

	e.obj.X0 = x0
	e.obj.Y0 = y0
	e.obj.Z0 = z0

	e.sprite = c
}
