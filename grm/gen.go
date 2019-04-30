package grm

import (
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"math/rand"
)

func GetAvgY(target *Obj) {
	avgtar := 0.0

	for i := 0; i < 16*16*4; i += 1 {
		avgtar += float64(target.Mesh.vertexData0[i*4 + 1])
	}

	avgtar /= 1024.0

	target.AvgY = avgtar
}

func Raise(target *Obj, diff float64) {
	for i := 0; i < 16*16*4; i += 1 {
		target.Mesh.vertexData0[i*4 + 1] += float32(diff)
	}
	target.CY += float64(diff)
}

func Smooth(target *Obj) {
	for i := 0; i < 16*16*4; i += 1 {
		y0 := target.Mesh.vertexData0[i*4 + 1]
		y0 = y0 - float32(target.AvgY)
		if math.Abs(float64(y0)) > 10.0 {
			target.Mesh.vertexData0[i*4 + 1] = float32(target.AvgY)
		}
	}
}

func Smooth0(target *Obj) {
	return
	for i := 0; i < 16*16*2 - 1; i += 1 {
		y0 := target.Mesh.vertexData0[i*8 + 1]
		y1 := target.Mesh.vertexData0[i*8 + 4 + 1]
		y0 = y0*0.8 + y1*0.2
		y1 = y1*0.8 + y0*0.2
		target.Mesh.vertexData0[i*8 + 1] = y0
		target.Mesh.vertexData0[(i + 1)*8 + 1] = y0
		target.Mesh.vertexData0[i*8 + 1] = y0
		target.Mesh.vertexData0[i*8 + 1] = y0
		target.Mesh.vertexData0[i*8 + 4 + 1] = y1
		target.Mesh.vertexData0[i*8 + 4 + 1] = y1
		target.Mesh.vertexData0[i*8 + 4 + 1] = y1
		target.Mesh.vertexData0[i*8 + 4 + 1] = y1
	}
}

func Smooth1(target *Obj) {
	return
	for i := 0; i < 15; i += 1 {
		y0 := target.Mesh.vertexData0[16 * 4 *4 * (i) + 1]
		y1 := target.Mesh.vertexData0[16 * 4 *4 * (i + 1) + 1]
		y0 = y0*0.8 + y1*0.2
		y1 = y1*0.8 + y0*0.2
		target.Mesh.vertexData0[16 * 4 *4 * (i) + 1] = y0
		target.Mesh.vertexData0[16 * 4 *4 * (i + 1) + 1] = y1
	}
}

func Flatten(target *Obj, template *Obj) {
	avgtar := 0.0
	avgtem := 0.0

	for i := 0; i < 16*16*4; i += 1 {
		avgtar += float64(target.Mesh.vertexData0[i*4 + 1])
		avgtem += float64(template.Mesh.vertexData0[i*4 + 1])
	}

	avgtar /= 1024.0
	avgtem /= 1024.0

	diff := float32(-avgtar + avgtem)

	for i := 0; i < 16*16*4; i += 1 {
		target.Mesh.vertexData0[i*4 + 1] += diff*0.5
	}
	target.Y0 += float64(diff*0.5)
}

func Stitch0(target *Obj, template *Obj) {
	//verts := len(target.vertexData0)
	row := 16 * 4 * 4

	for i := 0; i < 16; i += 1 {
		target.Mesh.vertexData0[0*4+0+i*4*4] = template.Mesh.vertexData0[row*15+1*4+0+i*4*4]
		target.Mesh.vertexData0[0*4+1+i*4*4] = template.Mesh.vertexData0[row*15+1*4+1+i*4*4]
		target.Mesh.vertexData0[0*4+2+i*4*4] = template.Mesh.vertexData0[row*15+1*4+2+i*4*4]
		target.Mesh.vertexData0[0*4+3+i*4*4] = template.Mesh.vertexData0[row*15+1*4+3+i*4*4]

		target.Mesh.vertexData0[2*4+4+i*4*4] = template.Mesh.vertexData0[row*15+2*4+0+i*4*4]
		target.Mesh.vertexData0[2*4+5+i*4*4] = template.Mesh.vertexData0[row*15+2*4+1+i*4*4]
		target.Mesh.vertexData0[2*4+6+i*4*4] = template.Mesh.vertexData0[row*15+2*4+2+i*4*4]
		target.Mesh.vertexData0[2*4+7+i*4*4] = template.Mesh.vertexData0[row*15+2*4+3+i*4*4]
	}
}

func Stitch1(target *Obj, template *Obj) {
	//verts := len(target.vertexData0)
	row := 16 * 4 * 4

	for i := 0; i < 16; i += 1 {
		target.Mesh.vertexData0[1*4+0+0*4*4+i*row] = template.Mesh.vertexData0[2*4+0+15*4*4+i*row]
		target.Mesh.vertexData0[1*4+1+0*4*4+i*row] = template.Mesh.vertexData0[2*4+1+15*4*4+i*row]
		target.Mesh.vertexData0[1*4+2+0*4*4+i*row] = template.Mesh.vertexData0[2*4+2+15*4*4+i*row]
		target.Mesh.vertexData0[1*4+3+0*4*4+i*row] = template.Mesh.vertexData0[2*4+3+15*4*4+i*row]

		target.Mesh.vertexData0[0*4+0+0*4*4+i*row] = template.Mesh.vertexData0[3*4+0+15*4*4+i*row]
		target.Mesh.vertexData0[0*4+1+0*4*4+i*row] = template.Mesh.vertexData0[3*4+1+15*4*4+i*row]
		target.Mesh.vertexData0[0*4+2+0*4*4+i*row] = template.Mesh.vertexData0[3*4+2+15*4*4+i*row]
		target.Mesh.vertexData0[0*4+3+0*4*4+i*row] = template.Mesh.vertexData0[3*4+3+15*4*4+i*row]
	}
}

func Stitch2(target *Obj, template *Obj) {
	//verts := len(target.vertexData0)
	row := 16 * 4 * 4

	for i := 0; i < 16; i += 1 {
		template.Mesh.vertexData0[row*15+1*4+0+i*4*4] = target.Mesh.vertexData0[0*4+0+i*4*4]
		template.Mesh.vertexData0[row*15+1*4+1+i*4*4] = target.Mesh.vertexData0[0*4+1+i*4*4]
		template.Mesh.vertexData0[row*15+1*4+2+i*4*4] = target.Mesh.vertexData0[0*4+2+i*4*4]
		template.Mesh.vertexData0[row*15+1*4+3+i*4*4] = target.Mesh.vertexData0[0*4+3+i*4*4]

		template.Mesh.vertexData0[row*15+2*4+0+i*4*4] = target.Mesh.vertexData0[2*4+4+i*4*4]
		template.Mesh.vertexData0[row*15+2*4+1+i*4*4] = target.Mesh.vertexData0[2*4+5+i*4*4]
		template.Mesh.vertexData0[row*15+2*4+2+i*4*4] = target.Mesh.vertexData0[2*4+6+i*4*4]
		template.Mesh.vertexData0[row*15+2*4+3+i*4*4] = target.Mesh.vertexData0[2*4+7+i*4*4]
	}
}

func Stitch3(target *Obj, template *Obj) {
	//verts := len(target.vertexData0)
	row := 16 * 4 * 4

	for i := 0; i < 16; i += 1 {
		template.Mesh.vertexData0[2*4+0+15*4*4+i*row] = target.Mesh.vertexData0[1*4+0+0*4*4+i*row]
		template.Mesh.vertexData0[2*4+1+15*4*4+i*row] = target.Mesh.vertexData0[1*4+1+0*4*4+i*row]
		template.Mesh.vertexData0[2*4+2+15*4*4+i*row] = target.Mesh.vertexData0[1*4+2+0*4*4+i*row]
		template.Mesh.vertexData0[2*4+3+15*4*4+i*row] = target.Mesh.vertexData0[1*4+3+0*4*4+i*row]

		template.Mesh.vertexData0[3*4+0+15*4*4+i*row] = target.Mesh.vertexData0[0*4+0+0*4*4+i*row]
		template.Mesh.vertexData0[3*4+1+15*4*4+i*row] = target.Mesh.vertexData0[0*4+1+0*4*4+i*row]
		template.Mesh.vertexData0[3*4+2+15*4*4+i*row] = target.Mesh.vertexData0[0*4+2+0*4*4+i*row]
		template.Mesh.vertexData0[3*4+3+15*4*4+i*row] = target.Mesh.vertexData0[0*4+3+0*4*4+i*row]
	}
}

func GenChunkDim(x0 float64, z0 float64, dimsq [][]int, dimSize int) Obj {
	var newChunk Mesh

	edgeLength := 2.0

	chunkSizeX := 16
	chunkSizeZ := 16

	//maxHeight := 20.0

	doffsi := int((x0) / float64(edgeLength))
	doffsj := int((z0) / float64(edgeLength))

	if doffsi > dimSize-1 {
		do := dimSize * int(math.Ceil(float64(doffsi)/float64(dimSize)))
		doffsi -= do
	}
	if doffsj > dimSize-1 {
		do := dimSize * int(math.Ceil(float64(doffsj)/float64(dimSize)))
		doffsj -= do
	}
	if doffsi < 0 {
		do := dimSize * int(math.Ceil(float64(-doffsi)/float64(dimSize)))
		doffsi += do
	}
	if doffsj < 0 {
		do := dimSize * int(math.Ceil(float64(-doffsj)/float64(dimSize)))
		doffsj += do
	}

	//x0 := 0.0
	y0 := float64(dimsq[doffsi][doffsj])

	//z0 := 0.0

	x := x0
	y := y0
	z := z0

	centerX := 0.0
	centerY := 0.0
	centerZ := 0.0

	t0 := 0.0
	t1 := 0.0

	t2 := 0.5
	t3 := 0.5

	yz0 := float64(dimsq[doffsi][doffsj])
	yza := make([]float64, chunkSizeZ)

	for i := 0; i < chunkSizeZ; i++ {
		yza[i] = y0
	}
	yz0a := make([]float64, chunkSizeZ)

	for i := 0; i < chunkSizeZ; i++ {
		yz0a[i] = y0
	}

	yz := yz0

	index := 0

	newChunk.clearMesh()

	newChunk.Model = mgl32.Ident4()

	var di int
	var dj int

	for i := 0; i < chunkSizeX; i++ {
		for j := 0; j < chunkSizeZ; j++ {
			di = i + doffsj
			dj = j + doffsi

			if di > dimSize-1 {
				di -= dimSize * int(di/dimSize)
			}
			if dj > dimSize-1 {
				dj -= dimSize * int(dj/dimSize)
			}

			yz = float64(dimsq[di][dj])

			if i == 0 && j == 0 {
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)
			} else if i == 0 {
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz0))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz0))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)
			} else if j == 0 {
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yza[j]))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yza[j]))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)
			} else {
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz0a[j]))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz0))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

				newChunk.vertexData0 = append(newChunk.vertexData0, float32(x))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yza[j]))
				newChunk.vertexData0 = append(newChunk.vertexData0, float32(z+edgeLength))
				newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)
			}

			if i == chunkSizeX/2-1 && j == chunkSizeZ/2-1 {
				centerX = x
				centerY = yz
				centerZ = z
			}

			var tr float64
			if rand.Float64() > 0.5 {
				tr = rand.Float64()
			}
			t0, t1, t2, t3 = randomTex((float64(int(yz)%40)/40.0 + tr) / 2.0 + 0.2)

			newChunk.texData = append(newChunk.texData, float32(t0))
			newChunk.texData = append(newChunk.texData, float32(t1))
			newChunk.texData = append(newChunk.texData, float32(1.0))
			newChunk.texData = append(newChunk.texData, float32(1.0))

			newChunk.texData = append(newChunk.texData, float32(t0+t2))
			newChunk.texData = append(newChunk.texData, float32(t1))
			newChunk.texData = append(newChunk.texData, float32(1.0))
			newChunk.texData = append(newChunk.texData, float32(1.0))

			newChunk.texData = append(newChunk.texData, float32(t0+t2))
			newChunk.texData = append(newChunk.texData, float32(t1+t3))
			newChunk.texData = append(newChunk.texData, float32(1.0))
			newChunk.texData = append(newChunk.texData, float32(1.0))

			newChunk.texData = append(newChunk.texData, float32(t0))
			newChunk.texData = append(newChunk.texData, float32(t1+t3))
			newChunk.texData = append(newChunk.texData, float32(1.0))
			newChunk.texData = append(newChunk.texData, float32(1.0))

			newChunk.indexData = append(newChunk.indexData, uint16(index+2))
			newChunk.indexData = append(newChunk.indexData, uint16(index+1))
			newChunk.indexData = append(newChunk.indexData, uint16(index+0))

			newChunk.indexData = append(newChunk.indexData, uint16(index+3))
			newChunk.indexData = append(newChunk.indexData, uint16(index+2))
			newChunk.indexData = append(newChunk.indexData, uint16(index+0))

			index += 4
			z += edgeLength

			yz0a[j] = yz0
			yza[j] = yz
			yz0 = yz
		}
		x += edgeLength
		z = z0
		yz0 = float64(dimsq[di][doffsj])
		yz = float64(dimsq[di][doffsj])
	}

	newChunk.vertexData = append(newChunk.vertexData, newChunk.vertexData0...)

	newChunk.faceNormals = append(newChunk.faceNormals, filledArray((int)((float64)(len(newChunk.indexData))/3.0)*3, 0.0)...)
	newChunk.faceCenter = append(newChunk.faceCenter, filledArray((int)((float64)(len(newChunk.indexData))/3.0)*3, 0.0)...)

	newChunk.vertexNormals = append(newChunk.vertexNormals, filledArray(len(newChunk.vertexData), 0.0)...)

	newChunk.Update()

	newChunk.getTriSize()

	var nc Obj
	//nc.LoadHullMesh(nil, newChunk, true, true, "")
	nc.Mesh = newChunk

	nc.X0 = x0
	nc.Y0 = y0
	nc.Z0 = z0

	nc.CX = centerX
	nc.CY = centerY
	nc.CZ = centerZ

	return nc
}

func GenChunkDimFlat(x0 float64, z0 float64, xs int, zs int, dimsq [][]int, dimSize int) Mesh {
	var newChunk Mesh

	edgeLength := 2.0

	chunkSizeX := 16
	chunkSizeZ := 16

	//maxHeight := 20.0

	doffsi := int(xs)
	doffsj := int(zs)

	//x0 := 0.0
	y0 := float64(dimsq[doffsi][doffsj])

	//z0 := 0.0

	x := x0
	y := y0
	z := z0

	t0 := 0.0
	t1 := 0.0

	t2 := 0.5
	t3 := 0.5

	yz0 := float64(dimsq[doffsi][doffsj])
	yza := make([]float64, chunkSizeZ)

	for i := 0; i < chunkSizeZ; i++ {
		yza[i] = y0
	}
	yz0a := make([]float64, chunkSizeZ)

	for i := 0; i < chunkSizeZ; i++ {
		yz0a[i] = y0
	}

	yz := yz0

	index := 0

	newChunk.clearMesh()

	newChunk.Model = mgl32.Ident4()

	for i := 0; i < chunkSizeX; i++ {
		for j := 0; j < chunkSizeZ; j++ {
			di := i + doffsj
			dj := j + doffsi

			if di > dimSize-1 {
				di -= dimSize * int(di/dimSize)
			}
			if dj > dimSize-1 {
				dj -= dimSize * int(dj/dimSize)
			}

			yz = float64(dimsq[di][dj])

			newChunk.vertexData0 = append(newChunk.vertexData0, float32(x))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(z))
			newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

			newChunk.vertexData0 = append(newChunk.vertexData0, float32(x+edgeLength))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(z))
			newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

			newChunk.vertexData0 = append(newChunk.vertexData0, float32(x+edgeLength))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(z+edgeLength))
			newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

			newChunk.vertexData0 = append(newChunk.vertexData0, float32(x))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(y*0.0+yz))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(z+edgeLength))
			newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

			t0, t1, t2, t3 = randomTex(yz)

			newChunk.texData = append(newChunk.texData, float32(t0))
			newChunk.texData = append(newChunk.texData, float32(t1))
			newChunk.texData = append(newChunk.texData, float32(1.0))
			newChunk.texData = append(newChunk.texData, float32(1.0))

			newChunk.texData = append(newChunk.texData, float32(t0+t2))
			newChunk.texData = append(newChunk.texData, float32(t1))
			newChunk.texData = append(newChunk.texData, float32(1.0))
			newChunk.texData = append(newChunk.texData, float32(1.0))

			newChunk.texData = append(newChunk.texData, float32(t0+t2))
			newChunk.texData = append(newChunk.texData, float32(t1+t3))
			newChunk.texData = append(newChunk.texData, float32(1.0))
			newChunk.texData = append(newChunk.texData, float32(1.0))

			newChunk.texData = append(newChunk.texData, float32(t0))
			newChunk.texData = append(newChunk.texData, float32(t1+t3))
			newChunk.texData = append(newChunk.texData, float32(1.0))
			newChunk.texData = append(newChunk.texData, float32(1.0))

			newChunk.indexData = append(newChunk.indexData, uint16(index+2))
			newChunk.indexData = append(newChunk.indexData, uint16(index+1))
			newChunk.indexData = append(newChunk.indexData, uint16(index+0))

			newChunk.indexData = append(newChunk.indexData, uint16(index+3))
			newChunk.indexData = append(newChunk.indexData, uint16(index+2))
			newChunk.indexData = append(newChunk.indexData, uint16(index+0))

			index += 4
			z += edgeLength

			yz0a[j] = yz0
			yza[j] = yz
			yz0 = yz
		}
		x += edgeLength
		z = z0
		yz0 = float64(dimsq[i][doffsj])
		yz = float64(dimsq[i][doffsj])
	}

	newChunk.vertexData = append(newChunk.vertexData, newChunk.vertexData0...)

	newChunk.faceNormals = append(newChunk.faceNormals, filledArray((int)((float64)(len(newChunk.indexData))/3.0)*3, 0.0)...)
	newChunk.faceCenter = append(newChunk.faceCenter, filledArray((int)((float64)(len(newChunk.indexData))/3.0)*3, 0.0)...)

	newChunk.vertexNormals = append(newChunk.vertexNormals, filledArray(len(newChunk.vertexData), 0.0)...)

	newChunk.Update()

	newChunk.getTriSize()

	return newChunk
}

func randomTex(seed float64) (float64, float64, float64, float64) {
	n := seed

	if n <= 0.25 {
		return 0.071, 1.0 - 0.650, 0.0122, 0.0122
	} else if n > 0.25 && n <= 0.5 {
		return 0.071, 1.0 - 0.861, 0.0122, 0.0122
	} else if n > 0.5 && n <= 0.75 {
		return 0.317, 1.0 - 0.873, 0.0122, 0.0122
	} else {
		return 0.293, 1.0 - 0.663, 0.0122, 0.0122
	}
}

func GenChunkFlat(x0 float64, z0 float64) Mesh {
	var newChunk Mesh

	edgeLength := 2.0

	chunkSizeX := 16
	chunkSizeZ := 16

	//x0 := 0.0
	y0 := 0.0
	//z0 := 0.0

	x := x0
	y := y0
	z := z0

	t0 := 0.75
	t1 := 0.75

	t2 := 0.05
	t3 := 0.05

	index := 0

	newChunk.clearMesh()

	newChunk.Model = mgl32.Ident4()

	for i := 0; i < chunkSizeX; i++ {
		for j := 0; j < chunkSizeZ; j++ {
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(x))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(y))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(z))
			newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

			newChunk.vertexData0 = append(newChunk.vertexData0, float32(x+edgeLength))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(y))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(z))
			newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

			newChunk.vertexData0 = append(newChunk.vertexData0, float32(x+edgeLength))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(y))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(z+edgeLength))
			newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

			newChunk.vertexData0 = append(newChunk.vertexData0, float32(x))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(y))
			newChunk.vertexData0 = append(newChunk.vertexData0, float32(z+edgeLength))
			newChunk.vertexData0 = append(newChunk.vertexData0, 1.0)

			t0, t1, t2, t3 = randomTex(0)

			newChunk.texData = append(newChunk.texData, float32(t0))
			newChunk.texData = append(newChunk.texData, float32(t1))
			newChunk.texData = append(newChunk.texData, float32(1.0))
			newChunk.texData = append(newChunk.texData, float32(1.0))

			newChunk.texData = append(newChunk.texData, float32(t0+t2))
			newChunk.texData = append(newChunk.texData, float32(t1))
			newChunk.texData = append(newChunk.texData, float32(1.0))
			newChunk.texData = append(newChunk.texData, float32(1.0))

			newChunk.texData = append(newChunk.texData, float32(t0+t2))
			newChunk.texData = append(newChunk.texData, float32(t1+t3))
			newChunk.texData = append(newChunk.texData, float32(1.0))
			newChunk.texData = append(newChunk.texData, float32(1.0))

			newChunk.texData = append(newChunk.texData, float32(t0))
			newChunk.texData = append(newChunk.texData, float32(t1+t3))
			newChunk.texData = append(newChunk.texData, float32(1.0))
			newChunk.texData = append(newChunk.texData, float32(1.0))

			newChunk.indexData = append(newChunk.indexData, uint16(index+2))
			newChunk.indexData = append(newChunk.indexData, uint16(index+1))
			newChunk.indexData = append(newChunk.indexData, uint16(index+0))

			newChunk.indexData = append(newChunk.indexData, uint16(index+3))
			newChunk.indexData = append(newChunk.indexData, uint16(index+2))
			newChunk.indexData = append(newChunk.indexData, uint16(index+0))

			index += 4
			z += edgeLength
		}
		x += edgeLength
		z = z0
	}

	newChunk.vertexData = append(newChunk.vertexData, newChunk.vertexData0...)

	newChunk.faceNormals = append(newChunk.faceNormals, filledArray((int)((float64)(len(newChunk.indexData))/3.0)*3, 0.0)...)
	newChunk.faceCenter = append(newChunk.faceCenter, filledArray((int)((float64)(len(newChunk.indexData))/3.0)*3, 0.0)...)

	newChunk.vertexNormals = append(newChunk.vertexNormals, filledArray(len(newChunk.vertexData), 0.0)...)

	newChunk.Update()

	newChunk.getTriSize()

	return newChunk
}

func DiamondSquare(h [][]int, size int) {
	zSize := len(h)
	xSize := len(h[0])

	half := size / 2

	if half < 1 {
		return
	}

	for z := half; z < zSize; z += size {
		for x := half; x < xSize; x += size {
			square(h, x%xSize, z%zSize, half)
		}
	}

	col := 0

	for x := 0; x < xSize; x += half {
		col++

		if col%2 == 1 {
			for z := half; z < zSize; z += size {
				diamond(h, x%xSize, z%zSize, half)
			}
		} else {
			for z := 0; z < zSize; z += size {
				diamond(h, x%xSize, z%zSize, half)
			}
		}
	}

	DiamondSquare(h, size/2)
}

func square(h [][]int, x int, z int, reach int) {
	zSize := len(h)
	xSize := len(h[0])
	count := 0

	avg := 0.0

	if x-reach >= 0 && z-reach >= 0 {
		avg += float64(h[x-reach][z-reach])
		count++
	}

	if x-reach >= 0 && z+reach < zSize {
		avg += float64(h[x-reach][z+reach])
		count++
	}

	if x+reach < xSize && z-reach >= 0 {
		avg += float64(h[x+reach][z-reach])
		count++
	}

	if x+reach < xSize && z+reach < zSize {
		avg += float64(h[x+reach][z+reach])
		count++
	}

	avg += rand.Float64() * float64(reach)

	avg /= float64(count)

	h[x][z] = int(math.Round(avg))
}

func diamond(h [][]int, x int, z int, reach int) {
	zSize := len(h)
	xSize := len(h[0])
	count := 0

	avg := 0.0

	if x-reach >= 0 {
		avg += float64(h[x-reach][z])
		count++
	}

	if x+reach < xSize {
		avg += float64(h[x+reach][z])
		count++
	}

	if z-reach >= 0 {
		avg += float64(h[x][z-reach])
		count++
	}

	if z+reach < zSize {
		avg += float64(h[x][z+reach])
		count++
	}

	avg += rand.Float64() * float64(reach)

	avg /= float64(count)

	h[x][z] = int(math.Floor(avg))
}
