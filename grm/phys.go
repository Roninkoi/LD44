package grm

import (
	"github.com/go-gl/mathgl/mgl32"
	"math"
)

var epsilon float32 = 0.01

type Phys struct {
	RPos mgl32.Vec3
	Pos  mgl32.Vec3
	Rot  mgl32.Vec3

	V     mgl32.Vec3 // velocity
	A     mgl32.Vec3 // acceleration
	force mgl32.Vec3

	av     mgl32.Vec3 // angular
	aa     mgl32.Vec3
	torque mgl32.Vec3

	Scale mgl32.Vec3

	mass float64

	IsStatic bool

	RenderPos bool
}

func Nv3() mgl32.Vec3 { // null vector
	return mgl32.Vec3{0.0, 0.0, 0.0}
}

func Nv4() mgl32.Vec4 {
	return mgl32.Vec4{0.0, 0.0, 0.0, 0.0}
}

func Iv3() mgl32.Vec3 { // one vector
	return mgl32.Vec3{1.0, 1.0, 1.0}
}

func (p *Phys) Init() {
	p.RPos = Nv3()
	p.Pos = Nv3()
	p.Rot = Nv3()

	p.V = Nv3()
	p.A = Nv3()
	p.force = Nv3()

	p.av = Nv3()
	p.aa = Nv3()
	p.torque = Nv3()

	p.Scale = Iv3()

	p.mass = 1.0

	p.IsStatic = false

	p.RenderPos = false
}

func (p *Phys) Update() {
	p.Pos = p.Pos.Add(p.V)

	if p.RenderPos {
		np := mgl32.Vec3{p.Pos[0], p.Pos[1], p.Pos[2]}
		np = np.Mul(0.4)
		p.RPos = p.RPos.Mul(0.6)
		p.RPos = p.RPos.Add(np)
	} else {
		p.RPos = p.Pos
	}
}

type Hull struct {
	mesh Mesh

	cn mgl32.Vec3 // normal center
	no bool       // normals outside?
}

func (h *Hull) loadHull(p string, no bool) {
	h.mesh.LoadMesh(nil, p, "")
	h.no = no
}

func (h *Hull) loadHullMesh(m Mesh, no bool) {
	h.mesh = m
	h.no = no
}

func (h *Hull) update(u *mgl32.Mat4) {
	h.mesh.Model = *u
	h.mesh.Update()
	h.mesh.getTriSize()
}

// general intersection
func (h *Hull) intersectsMesh(i *Mesh) bool {
	returns := false

	cni := 0

	cnold := h.cn
	h.cn = mgl32.Vec3{0.0, 0.0, 0.0}

	for j := 0; j < (int)((float64)(len(h.mesh.faceCenter))/3.0) && !returns; j++ {
		for k := 0; k < (int)((float64)(len(i.faceCenter))/3.0) && !returns; k++ {
			a := mgl32.Vec3{h.mesh.faceCenter[j*3+0], h.mesh.faceCenter[j*3+1], h.mesh.faceCenter[j*3+2]}
			b := mgl32.Vec3{i.faceCenter[k*3+0], i.faceCenter[k*3+1], i.faceCenter[k*3+2]}

			c := a.Sub(b)
			c = c.Normalize()

			n := mgl32.Vec3{h.mesh.faceNormals[j*3+0], h.mesh.faceNormals[j*3+1], h.mesh.faceNormals[j*3+2]}

			cdn := c.Dot(n)

			is := cdn < 0.0

			if h.no {
				is = cdn > 0.0
			}

			if is {
				returns = true
				cni += 1

				h.cn = h.cn.Add(n)
				h.cn = h.cn.Mul(a.Sub(b).Len()*0.02 + 1.0)
			}
		}
	}

	if returns {
		h.cn = h.cn.Mul(1.0 / (float32)(cni))
	}
	if h.cn.Len() == 0.0 {
		h.cn = cnold
	}

	if !h.no {
		//h.cn = h.cn.Mul(-1.0)
	}

	return returns
}

// sphere intersection
func (h *Hull) intersectsSphere(sc mgl32.Vec3, sr float32) (bool, float64) {
	returns := false

	cni := 0

	h.cn = mgl32.Vec3{0.0, 0.0, 0.0}

	cd := 10.0

	for j := 0; j < (int)((float64)(len(h.mesh.faceCenter))/3.0) && !returns; j++ { // when 1, exit, might break
		a := mgl32.Vec3{h.mesh.faceCenter[j*3+0], h.mesh.faceCenter[j*3+1], h.mesh.faceCenter[j*3+2]}
		b := mgl32.Vec3{sc[0], sc[1], sc[2]}

		n := mgl32.Vec3{h.mesh.faceNormals[j*3+0], h.mesh.faceNormals[j*3+1], h.mesh.faceNormals[j*3+2]}

		b = b.Add(n.Mul(sr))

		c := a.Sub(b)

		if float64(c.Len()) < cd {
			cd = float64(c.Len())
		}

		if c.Len() < h.mesh.triSize*1.0 {
			cdn := c.Dot(n)

			if cdn < 0.0 {
				returns = true
				cni += 1

				h.cn = h.cn.Add(n)
			}
		}
	}

	if returns {
		h.cn = h.cn.Mul(1.0 / (float32)(cni))
	}

	if h.no {
		h.cn = h.cn.Mul(-1.0)
	}

	return returns, cd
}

type Sys struct {
	objs  []*Obj
	field mgl32.Vec3

	ticks float64
}

func (p *Sys) Init() {
	p.field = mgl32.Vec3{0.0, 0.01, 0.0}
}

func (p *Sys) Clear() {
	p.objs = nil
}

func (p *Sys) Add(o *Obj) {
	p.objs = append(p.objs, o)
}

func (p *Sys) physIsect(i int, j int) {
	isects := false

	if p.objs[j].SphereIsect {
		isects, p.objs[j].IDist = p.objs[i].Hull.intersectsSphere(p.objs[j].Mesh.SphereCenter, p.objs[j].Mesh.SphereRadius)
	} else {
		isects = p.objs[i].Hull.intersectsMesh(&p.objs[j].Mesh)
	}

	if isects {
		p.objs[i].Isects = true
		p.objs[j].Isects = true

		p.objs[j].Hull.cn = p.objs[i].Hull.cn.Mul(-1.0)
		nv := p.objs[i].Hull.cn.Normalize()

		dynamic := !p.objs[i].Phys.IsStatic && !p.objs[j].Phys.IsStatic

		if dynamic {
			p.objs[i].Phys.V = p.objs[i].Phys.V.Mul(0.5)
			p.objs[j].Phys.V = p.objs[j].Phys.V.Mul(0.5)
		}

		v1 := p.objs[i].Phys.V
		v2 := p.objs[j].Phys.V

		v := v2.Sub(v1)

		v = v.Add(p.field)

		vl := v.Len()

		vl *= float32(math.Abs(math.Abs(float64(nv.Dot(v.Normalize())))))

		if vl > 10.0*p.field.Len() {
			vl = 10.0*p.field.Len()
		}

		if math.IsNaN(float64(nv.Len())) {
			p.objs[i].Hull.cn = Nv3()
		}

		if math.IsNaN(float64(vl)) {
			return
		}

		if !p.objs[i].Phys.IsStatic {
			diff := nv.Mul(vl)
			p.objs[i].Phys.Pos = p.objs[i].Phys.Pos.Add(diff)
			p.objs[i].Phys.V = p.objs[i].Phys.V.Add(diff).Mul(0.9)
		}
		if !p.objs[j].Phys.IsStatic {
			diff := nv.Mul(-vl)
			p.objs[j].Phys.Pos = p.objs[j].Phys.Pos.Add(diff)
			p.objs[j].Phys.V = p.objs[j].Phys.V.Add(diff).Mul(0.9)
		}
	}
}

func (p *Sys) Update() {
	for i := 0; i < len(p.objs); i++ {
		p.objs[i].Isects = false
		for j := i + 1; j < len(p.objs); j++ {
			if (!p.objs[i].Phys.IsStatic || !p.objs[j].Phys.IsStatic) &&
				(p.objs[i].InRange || p.objs[j].InRange) {
				if p.objs[i].Mesh.SphereIsect(&p.objs[j].Mesh) {
					if p.objs[i].Phys.V.Len() >= epsilon || p.objs[j].Phys.V.Len() >= epsilon {
						if p.objs[i].HasHull {
							p.physIsect(i, j)
						} else if p.objs[j].HasHull {
							p.physIsect(j, i)
						}
					}
				}
			}
		}
	}
	for i := 0; i < len(p.objs); i++ {
		if !p.objs[i].Phys.IsStatic && p.objs[i].InRange && p.objs[i].Phys.V.Len() >= epsilon {
			if p.objs[i].Phys.V.Len() < epsilon {
				p.objs[i].Phys.V = Nv3()
			}

			p.objs[i].Phys.V = p.objs[i].Phys.V.Add(p.field)
			p.objs[i].Phys.V = p.objs[i].Phys.V.Mul(0.99)
			p.objs[i].Update()

			if p.objs[i].Phys.V.Len() < epsilon { // don't sleep in midair
				p.objs[i].Phys.V = p.objs[i].Phys.V.Add(p.field.Normalize().Mul(epsilon*2.0)) // WAKE UP!
			}
		}
	}
}
