package grm

import (
	"github.com/go-gl/mathgl/mgl32"
	"strconv"
)

type Mesh struct {
	vertexData0 []float32
	vertexData []float32
	texData []float32
	colData []float32
	indexData []uint16

	vertexNormals []float32
	faceNormals []float32
	faceCenter []float32

	tex Texture

	Model mgl32.Mat4

	SphereCenter mgl32.Vec3
	SphereRadius float32

	triSize float32
}

func (m *Mesh) SetCol(c mgl32.Vec4) {
	m.colData = make([]float32, len(m.vertexData0))
	for i := 0; i < len(m.colData)/4; i += 1 {
		m.colData[i*4+0] = c[0]
		m.colData[i*4+1] = c[1]
		m.colData[i*4+2] = c[2]
		m.colData[i*4+3] = c[3]
	}
}

func (m *Mesh) EnableTransform() {
	for i := 0; i < len(m.vertexData0) / 4; i++ {
		m.vertexData0[i*4 + 3] = 1.0
	}
}

func (m *Mesh) DisableTransform() {
	for i := 0; i < len(m.vertexData0) / 4; i++ {
		m.vertexData0[i*4 + 3] = 0.0
	}
}

func parse(s string) []string {
	var returns []string

	word := ""
	for i := 0; i < len(s); i++ {
		if s[i] != ' ' && s[i] != '\n' && s[i] != '\r' {
			word += string(s[i])
		} else {
			returns = append(returns, word)
			word = ""
		}
	}

	return returns
}

func (m *Mesh) parseFace(s string) (f int, t int) {
	face := ""
	tex := ""

	texstart := false
	normstart := false
	for i := 0; i < len(s); i++ {
		if texstart {
			if s[i] != '/' && !normstart {
				tex += string(s[i])
			} else {
				normstart = true
			}
		}

		if s[i] != '/' && !texstart {
			face += string(s[i])
		} else {
			texstart = true
		}
	}

	facei, _ := strconv.Atoi(face)
	texi, _ := strconv.Atoi(tex)

	return facei - 1, texi - 1 // obj indices start at 1
}

func (m *Mesh) clearMesh() {
	m.vertexData = nil
	m.vertexData0 = nil
	m.texData = nil
	m.indexData = nil
	m.faceNormals = nil
	m.faceCenter = nil
	m.colData = nil
}

func (m *Mesh) LoadMesh(r *Renderer, p string, t string) {
	if len(t) > 0 {
		m.tex.Load(r, t)
	} else {
		m.tex.Path = ""
	}

	s := readFileString(p)

	sa := parse(s)

	m.clearMesh()

	m.Model = mgl32.Ident4()

	var vd []float32
	var td []float32
	var vfd []uint16
	var tfd []uint16

	for i := 0; i < len(sa); i++ {
		if sa[i] == "v" {
			if len(sa) > i+3 {
				vx, _ := strconv.ParseFloat(sa[i+1], 32)
				vy, _ := strconv.ParseFloat(sa[i+2], 32)
				vy = -vy // flip y for vulkan
				vz, _ := strconv.ParseFloat(sa[i+3], 32)

				vd = append(vd, (float32)(vx))
				vd = append(vd, (float32)(vy))
				vd = append(vd, (float32)(vz))
				vd = append(vd, 1.0)
			}
		}
		if sa[i] == "vt" {
			if len(sa) > i+2 {
				tx, _ := strconv.ParseFloat(sa[i+1], 32)
				ty, _ := strconv.ParseFloat(sa[i+2], 32)

				td = append(td, (float32)(tx))
				td = append(td, (float32)(ty))
				td = append(td, 1.0)
				td = append(td, 1.0)
			}
		}
		if sa[i] == "f" {
			if len(sa) > i+3 {
				f0, t0 := m.parseFace(sa[i+1])
				f1, t1 := m.parseFace(sa[i+2])
				f2, t2 := m.parseFace(sa[i+3])

				vfd = append(vfd, (uint16)(f0))
				vfd = append(vfd, (uint16)(f1))
				vfd = append(vfd, (uint16)(f2))

				tfd = append(tfd, (uint16)(t0))
				tfd = append(tfd, (uint16)(t1))
				tfd = append(tfd, (uint16)(t2))
			}
		}
	}

	// mesh optimization

	var vpairs []uint16
	var tpairs []uint16

	for i := 0; i < len(vfd); i++ {
		exists := false

		j := 0
		for ; j < len(vpairs) && !exists; j++ {
			if vfd[i] == vpairs[j] && tfd[i] == tpairs[j] {
				exists = true
			}
		}

		if exists {
			m.indexData = append(m.indexData, (uint16)(j-1))
		} else {
			ei := j

			vpairs = append(vpairs, vfd[i])
			tpairs = append(tpairs, tfd[i])

			m.indexData = append(m.indexData, (uint16)(ei))

			m.vertexData0 = append(m.vertexData0, vd[vfd[i]*4])
			m.vertexData0 = append(m.vertexData0, vd[vfd[i]*4 + 1])
			m.vertexData0 = append(m.vertexData0, vd[vfd[i]*4 + 2])
			m.vertexData0 = append(m.vertexData0, 1.0)

			m.texData = append(m.texData, td[tfd[i]*4])
			m.texData = append(m.texData, td[tfd[i]*4 + 1])
			m.texData = append(m.texData, 1.0)
			m.texData = append(m.texData, 1.0)
		}
	}

	m.vertexData = append(m.vertexData, m.vertexData0...) // copy vertex data for transform

	m.faceNormals = append(m.faceNormals, filledArray((int)((float64)(len(m.indexData))/3.0)*3, 0.0)...)
	m.faceCenter = append(m.faceCenter, filledArray((int)((float64)(len(m.indexData))/3.0)*3, 0.0)...)

	m.vertexNormals = append(m.vertexNormals, filledArray(len(m.vertexData), 0.0)...)

	m.SetCol(mgl32.Vec4{1.0, 1.0, 1.0, 1.0})

	m.Update()

	m.getTriSize()
}

func (m *Mesh) loadVerts(p string) {
	s := readFileString(p)

	sa := parse(s)

	m.clearMesh()

	m.Model = mgl32.Ident4()

	for i := 0; i < len(sa); i++ {
		if sa[i] == "v" {
			if len(sa) > i+3 {
				vx, _ := strconv.ParseFloat(sa[i+1], 32)
				vy, _ := strconv.ParseFloat(sa[i+2], 32)
				vz, _ := strconv.ParseFloat(sa[i+3], 32)

				m.vertexData0 = append(m.vertexData0, (float32)(vx))
				m.vertexData0 = append(m.vertexData0, (float32)(vy))
				m.vertexData0 = append(m.vertexData0, (float32)(vz))
				m.vertexData0 = append(m.vertexData0, 1.0)
			}
		}
	}
}

func filledArray(l int, v float32) []float32 {
	var a []float32

	for i := 0; i <  l; i++ {
		a = append(a, v)
	}

	return a
}

func (m *Mesh) transformVerts() {
	for i := 0; i < len(m.vertexData); i += 4 {
		rv := mgl32.Vec3{m.vertexData0[i+0], m.vertexData0[i+1], m.vertexData0[i+2]}

		rv = mgl32.TransformCoordinate(rv, m.Model)

		m.vertexData[0+i] = rv[0]
		m.vertexData[1+i] = rv[1]
		m.vertexData[2+i] = rv[2]
		m.vertexData[3+i] = m.vertexData0[i+3]
	}
}

/*
 * just face normals
 */
func (m *Mesh) getNormals() {
	for i := 0; i < len(m.faceNormals)/3.0; i++ {
		v0 := mgl32.Vec3{m.vertexData[m.indexData[i*3+0]*4], m.vertexData[m.indexData[i*3+0]*4+1], m.vertexData[m.indexData[i*3+0]*4+2]}
		v1 := mgl32.Vec3{m.vertexData[m.indexData[i*3+1]*4], m.vertexData[m.indexData[i*3+1]*4+1], m.vertexData[m.indexData[i*3+1]*4+2]}
		v2 := mgl32.Vec3{m.vertexData[m.indexData[i*3+2]*4], m.vertexData[m.indexData[i*3+2]*4+1], m.vertexData[m.indexData[i*3+2]*4+2]}

		c := mgl32.Vec3{0.0, 0.0, 0.0}

		c = c.Add(v0)
		c = c.Add(v1)
		c = c.Add(v2)

		c = c.Mul(1.0 / 3.0)

		m.faceCenter[i*3 + 0] = c[0]
		m.faceCenter[i*3 + 1] = c[1]
		m.faceCenter[i*3 + 2] = c[2]

		a := v0.Sub(v1)
		b := v2.Sub(v1)

		axb := a.Cross(b)

		axb = axb.Normalize()

		m.faceNormals[i*3 + 0] = axb[0]
		m.faceNormals[i*3 + 1] = axb[1]
		m.faceNormals[i*3 + 2] = axb[2]
	}
}

func (m *Mesh) Update() {
	m.transformVerts()

	m.getNormals()

	m.getBoundingSphere()
}

func (m *Mesh) SphereIsect(n *Mesh) bool {
	intersects := false

	if m.SphereCenter.Sub(n.SphereCenter).Len() <= m.SphereRadius + n.SphereRadius {
		intersects = true
	}

	return intersects
}

func (m *Mesh) getBoundingSphere() {
	vn := 0
	m.SphereRadius = 0.0
	for i := 0; i < len(m.vertexData); i += 4 {
		vp := mgl32.Vec3{m.vertexData[i], m.vertexData[i+1], m.vertexData[i+2]}
		m.SphereCenter = m.SphereCenter.Add(vp)
		vn += 1
	}
	m.SphereCenter = m.SphereCenter.Mul(1.0 / (float32)(vn))

	for i := 0; i < len(m.vertexData); i += 4 {
		vp := mgl32.Vec3{m.vertexData[i], m.vertexData[i+1], m.vertexData[i+2]}
		nr := vp.Sub(m.SphereCenter).Len()
		if nr > m.SphereRadius {
			m.SphereRadius = nr
		}
	}
}

func (m *Mesh) getTriSize() {
	m.triSize = 0.0
	for i := 0; i < len(m.indexData)/3.0; i++ {
		v0 := mgl32.Vec3{m.vertexData[m.indexData[i*3 + 0]*4+0],m.vertexData[m.indexData[i*3 + 0]*4+1], m.vertexData[m.indexData[i*3 + 0]*4+2]}
		v1 := mgl32.Vec3{m.vertexData[m.indexData[i*3 + 1]*4+0],m.vertexData[m.indexData[i*3 + 1]*4+1], m.vertexData[m.indexData[i*3 + 1]*4+2]}
		v2 := mgl32.Vec3{m.vertexData[m.indexData[i*3 + 2]*4+0],m.vertexData[m.indexData[i*3 + 2]*4+1], m.vertexData[m.indexData[i*3 + 2]*4+2]}

		vl := (float32)(0.0)

		if v0.Sub(v1).Len() > vl {
			vl = v0.Sub(v1).Len()
		}
		if v1.Sub(v2).Len() > vl {
			vl = v1.Sub(v2).Len()
		}
		if v2.Sub(v0).Len() > vl {
			vl = v2.Sub(v0).Len()
		}

		if vl > m.triSize {
			m.triSize = vl
		}
	}
}

func (m *Mesh) Draw(r *Renderer) {
	r.Add(&m.vertexData, &m.texData, &m.colData, &m.indexData, &m.tex)
}
