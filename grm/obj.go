package grm

import "github.com/go-gl/mathgl/mgl32"

type Obj struct {
	Mesh Mesh
	Hull Hull

	Phys Phys

	Isects      bool
	IsectNormal mgl32.Vec3
	IDist float64

	HasHull  bool
	sameHull bool // hull is Mesh

	SphereIsect bool // using sphere intersection

	Removed bool

	InRange bool

	X0 float64
	Y0 float64
	Z0 float64

	CX float64
	CY float64
	CZ float64

	AvgY float64
}

func (o *Obj) Init() {
	o.Removed = false
	o.InRange = true
	o.Isects = false
	o.Phys.Init()
}

func (o *Obj) Update() {
	o.Phys.Update()
	model := mgl32.Ident4()

	model = model.Mul4(mgl32.Translate3D(o.Phys.RPos[0], o.Phys.RPos[1], o.Phys.RPos[2]))
	//model = model.Mul4(mgl32.Translate3D(o.Phys.Pos[0], o.Phys.Pos[1], o.Phys.Pos[2]))

	//model = model.Mul4(mgl32.HomogRotate3DX(o.Phys.Rot[0]))
	model = model.Mul4(mgl32.HomogRotate3DY(o.Phys.Rot[1]))
	//model = model.Mul4(mgl32.HomogRotate3DZ(o.Phys.Rot[2]))

	//model = model.Mul4(mgl32.Scale3D(o.Phys.Scale[0], o.Phys.Scale[1], o.Phys.Scale[2]))

	o.Mesh.Model = model
	o.Mesh.Update()

	o.IsectNormal = o.Hull.cn

	if o.HasHull {
		if o.sameHull {
			o.Hull.mesh = o.Mesh
		} else {
			o.Hull.update(&model)
		}
	}
}

func (o *Obj) LoadObj(r *Renderer, p string, t string) {
	o.Init()
	o.Mesh.LoadMesh(r, p, t)

	o.SphereIsect = false
	o.HasHull = false
	o.sameHull = false
	o.Removed = false

	o.Phys.Init()

	o.Update()
}

func MakeObj(r *Renderer, p string, t string) Obj {
	var o Obj

	o.LoadObj(r, p, t)

	return o
}

func (o *Obj) LoadHull(r *Renderer, p string, h string, no bool, s bool, t string) {
	o.Init()
	o.Mesh.LoadMesh(r, p, t)

	o.SphereIsect = false
	o.HasHull = false
	o.Removed = false

	if h == "0" {
		o.Hull.loadHull(p, no)
		o.HasHull = true
		o.sameHull = true
	} else {
		if h != "" {
			o.Hull.loadHull(h, no)
			o.HasHull = true
			o.sameHull = false
		}
	}

	o.Phys.Init()

	o.Phys.IsStatic = s

	if !s {
		o.Phys.V[1] = -0.01 // kick
	}

	o.Update()
}

func (o *Obj) LoadHullMesh(r *Renderer, m Mesh, no bool, s bool, t string) {
	o.Init()
	o.Mesh = m

	o.SphereIsect = false
	o.HasHull = false
	o.Removed = false

	o.Hull.loadHullMesh(m, no)
	o.HasHull = true
	o.sameHull = true

	o.Phys.Init()

	o.Phys.IsStatic = s

	if !s {
		o.Phys.V[1] = -0.01 // kick
	}

	o.Update()
}

func (o *Obj) Draw(r *Renderer) {
	o.Mesh.Draw(r)
}
