package game

import (
	"../grm"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"math/rand"
)

type World struct {
	entities []Entity
	objs     []grm.Obj
	environs []grm.Obj

	money []grm.Obj
	souls []grm.Obj

	chunks []grm.Obj

	player Entity

	sys grm.Sys // physics system

	far float32

	time float64

	dimsq   [][]int
	dimSize int

	cw int
	ch int

	soulsNum  int
	moneyNum  int
	quotaNum  int
	billNum   int
	salaryNum int
	shiftNum  int
	day       int

	soulTicks float64
}

func (w *World) load(r *grm.Renderer) {
	w.far = 100.0

	w.sys.Init()

	w.restart()

	w.player.obj.HasHull = false
	w.player.obj.SphereIsect = true
	w.player.obj.Init()
	w.player.obj.Phys.RenderPos = true
	w.player.obj.Phys.Pos[0] = 21.3
	w.player.obj.Phys.Pos[1] = 32.0
	w.player.obj.Phys.Pos[2] = 56.3
	w.player.obj.Phys.V[1] = -0.07
	w.player.obj.Update()

	w.dimSize = 257

	w.dimsq = make([][]int, w.dimSize)
	for z := 0; z < w.dimSize; z++ {
		w.dimsq[z] = make([]int, w.dimSize)
		for x := 0; x < w.dimSize; x++ {
			//dimsq[z][x] = int(rand.Float64() * maxHeight)
			w.dimsq[z][x] = 0
		}
	}

	w.dimsq[128][128] = 20 // diamond seed
	w.dimsq[0][0] = 11
	w.dimsq[0][256] = 15
	w.dimsq[256][0] = 8
	w.dimsq[256][0] = 19

	grm.DiamondSquare(w.dimsq, w.dimSize)

	w.cw = 4
	w.ch = 4

	for j := 0; j < w.ch; j++ { // generate starting position
		for i := 0; i < w.cw; i++ {
			cx := 32.0 * float64(i)
			cz := 32.0 * float64(j)

			nc := grm.GenChunkDimS(cx, cz)

			avg := 0.0
			n := 0
			prevx := w.findChunk(cx+32.0*1, cz)
			if prevx >= 0 {
				grm.GetAvgY(&w.chunks[prevx])
				avg += w.chunks[prevx].AvgY
				n++
			}
			prevz := w.findChunk(cx, cz+32.0*1)
			if prevz >= 0 {
				grm.GetAvgY(&w.chunks[prevz])
				avg += w.chunks[prevz].AvgY
				n++
			}
			prevx = w.findChunk(cx-32.0*1, cz)
			if prevx >= 0 {
				grm.GetAvgY(&w.chunks[prevx])
				avg += w.chunks[prevx].AvgY
				n++
			}
			prevz = w.findChunk(cx, cz-32.0*1)
			if prevz >= 0 {
				grm.GetAvgY(&w.chunks[prevz])
				avg += w.chunks[prevz].AvgY
				n++
			}
			grm.GetAvgY(&nc)

			avg = avg/float64(n) - nc.AvgY*0.95

			grm.Smooth(&nc)
			if n > 0 {
				grm.Raise(&nc, avg)
			}

			prevx = w.findChunk(cx+32.0*1, cz)
			if prevx >= 0 {
				grm.Stitch2(&w.chunks[prevx], &nc)
			}
			prevz = w.findChunk(cx, cz+32.0*1)
			if prevz >= 0 {
				grm.Stitch3(&w.chunks[prevz], &nc)
			}
			prevx = w.findChunk(cx-32.0*1, cz)
			if prevx >= 0 {
				grm.Stitch0(&nc, &w.chunks[prevx])
			}
			prevz = w.findChunk(cx, cz-32.0*1)
			if prevz >= 0 {
				grm.Stitch1(&nc, &w.chunks[prevz])
			}

			w.addChunk(nc)
		}
	}

	w.sys.Clear()
	w.addPhys()

	w.player.obj.Phys.Pos[1] = float32(w.chunks[0].AvgY - 1.0)
}

func (w *World) genChunkObjs(c *grm.Obj) {
	if rand.Float64() > 0.5 {
		var e1 Entity

		e1.load()
		e1.randomChar(c.CX, c.CY-4.0, c.CZ)

		w.addEntity(e1)
	}

	for rand.Float64() > 0.2 {
		var e2 grm.Obj
		e2.LoadObj(nil, "gfx/plant.obj", "")
		e2.Phys.Pos = mgl32.Vec3{float32(c.CX + (rand.Float64()-0.5)*20.0), float32(c.CY), float32(c.CZ + (rand.Float64()-0.5)*20.0)}
		e2.Update()

		w.addEnv(e2)
	}
	if rand.Float64() > 0.2 {
		var e2 grm.Obj
		e2.LoadObj(nil, "gfx/rock.obj", "")
		e2.Phys.Pos = mgl32.Vec3{float32(c.CX + (rand.Float64()-0.5)*20.0), float32(c.CY), float32(c.CZ + (rand.Float64()-0.5)*20.0)}
		e2.Update()

		w.addEnv(e2)
	}
	for rand.Float64() > 0.5 {
		var e2 grm.Obj
		e2.LoadObj(nil, "gfx/tree.obj", "")
		e2.Phys.Pos = mgl32.Vec3{float32(c.CX + (rand.Float64()-0.5)*20.0), float32(c.CY), float32(c.CZ + (rand.Float64()-0.5)*20.0)}
		e2.Update()

		w.addEnv(e2)
	}
	if rand.Float64() > 0.2 {
		var e2 grm.Obj
		e2.LoadObj(nil, "gfx/money.obj", "")
		e2.Phys.Pos = mgl32.Vec3{float32(c.CX + (rand.Float64()-0.5)*20.0), float32(c.CY) - 2.0, float32(c.CZ + (rand.Float64()-0.5)*20.0)}
		e2.Update()

		w.money = append(w.money, e2)
	}
}

func (w *World) addChunk(c grm.Obj) {
	c.LoadHullMesh(nil, c.Mesh, true, true, "")
	w.genChunkObjs(&c)
	w.chunks = append(w.chunks, c)

	w.sys.Clear()
	w.addPhys()
}

func (w *World) addPhys() {
	for i := 0; i < len(w.chunks); i++ {
		w.sys.Add(&w.chunks[i])
	}
	for i := 0; i < len(w.objs); i++ {
		w.sys.Add(&w.objs[i])
	}
	for i := 0; i < len(w.entities); i++ {
		w.sys.Add(&w.entities[i].obj)
	}
	for i := 0; i < len(w.souls); i++ {
		w.sys.Add(&w.souls[i])
	}
	w.sys.Add(&w.player.obj)
}

func (w *World) addObj(o grm.Obj) {
	w.objs = append(w.objs, o)
}

func (w *World) addEnv(o grm.Obj) {
	w.environs = append(w.environs, o)
}

func (w *World) addEntity(e Entity) {
	w.entities = append(w.entities, e)
}

func (w *World) findChunk(cx float64, cz float64) int {
	for i := 0; i < len(w.chunks); i++ {
		if w.chunks[i].X0 == cx && w.chunks[i].Z0 == cz {
			return i
		}
	}
	for i := 0; i < len(w.chunks); i++ {
		if math.Round(w.chunks[i].X0) == math.Round(cx) && w.chunks[i].Z0 == math.Round(cz) {
			return i
		}
	}
	return -1
}

func (w *World) findFree(cx float64, cz float64) int {
	max := 0.0
	maxi := -1
	for i := 0; i < len(w.chunks); i++ {
		x := w.chunks[i].X0 - cx
		z := w.chunks[i].Z0 - cz

		dist := math.Sqrt(x*x + z*z)
		if dist > max {
			max = dist
			maxi = i
		}
	}
	return maxi
}

func (w *World) chunkGen(cx float64, cz float64) {
	ind := w.findFree(cx, cz) //cx+32.0*4,

	if ind >= 0 {
		w.chunks[ind] = grm.GenChunkDimS(cx, cz/*, w.dimsq, w.dimSize*/)

		avg := 0.0
		n := 0
		prevx := w.findChunk(cx+32.0*1, cz)
		if prevx >= 0 {
			grm.GetAvgY(&w.chunks[prevx])
			avg += w.chunks[prevx].AvgY
			n++
		}
		prevz := w.findChunk(cx, cz+32.0*1)
		if prevz >= 0 {
			grm.GetAvgY(&w.chunks[prevz])
			avg += w.chunks[prevz].AvgY
			n++
		}
		prevx = w.findChunk(cx-32.0*1, cz)
		if prevx >= 0 {
			grm.GetAvgY(&w.chunks[prevx])
			avg += w.chunks[prevx].AvgY
			n++
		}
		prevz = w.findChunk(cx, cz-32.0*1)
		if prevz >= 0 {
			grm.GetAvgY(&w.chunks[prevz])
			avg += w.chunks[prevz].AvgY
			n++
		}
		grm.GetAvgY(&w.chunks[ind])

		avg = avg/float64(n) - w.chunks[ind].AvgY*0.95

		grm.Smooth(&w.chunks[ind])
		if n > 0 {
			grm.Raise(&w.chunks[ind], avg)
		}

		prevx = w.findChunk(cx+32.0*1, cz)
		if prevx >= 0 {
			grm.Stitch2(&w.chunks[prevx], &w.chunks[ind])
		}
		prevz = w.findChunk(cx, cz+32.0*1)
		if prevz >= 0 {
			grm.Stitch3(&w.chunks[prevz], &w.chunks[ind])
		}
		prevx = w.findChunk(cx-32.0*1, cz)
		if prevx >= 0 {
			grm.Stitch0(&w.chunks[ind], &w.chunks[prevx])
		}
		prevz = w.findChunk(cx, cz-32.0*1)
		if prevz >= 0 {
			grm.Stitch1(&w.chunks[ind], &w.chunks[prevz])
		}

		w.genChunkObjs(&w.chunks[ind])

		w.sys.Clear()
		w.addPhys()

		w.chunks[ind].LoadHullMesh(nil, w.chunks[ind].Mesh, true, true, "")
	}
}

func (w *World) autoGen() {
	var cx float64
	var cz float64
	for ci := 1; ci < 2; ci++ {
		cx = math.Floor(float64(w.player.pos[0])/32.0)*32.0 + 32.0*float64(ci)
		cz = math.Floor(float64(w.player.pos[2])/32.0) * 32.0

		if w.findChunk(cx, cz) < 0 {
			w.chunkGen(cx, cz)

		}
		cx = math.Floor(float64(w.player.pos[0])/32.0)*32.0 - 32.0*float64(ci)
		cz = math.Floor(float64(w.player.pos[2])/32.0) * 32.0

		if w.findChunk(cx, cz) < 0 {
			w.chunkGen(cx, cz)
		}

		cx = math.Floor(float64(w.player.pos[0])/32.0) * 32.0
		cz = math.Floor(float64(w.player.pos[2])/32.0)*32.0 + 32.0*float64(ci)

		if w.findChunk(cx, cz) < 0 {
			w.chunkGen(cx, cz)
		}

		cx = math.Floor(float64(w.player.pos[0])/32.0) * 32.0
		cz = math.Floor(float64(w.player.pos[2])/32.0)*32.0 - 32.0*float64(ci)

		if w.findChunk(cx, cz) < 0 {
			w.chunkGen(cx, cz)
		}

		// CORNERS

		cx = math.Floor(float64(w.player.pos[0])/32.0)*32.0 + 32.0*float64(ci)
		cz = math.Floor(float64(w.player.pos[2])/32.0)*32.0 - 32.0*float64(ci)

		if w.findChunk(cx, cz) < 0 {
			w.chunkGen(cx, cz)
		}
		cx = math.Floor(float64(w.player.pos[0])/32.0)*32.0 - 32.0*float64(ci)
		cz = math.Floor(float64(w.player.pos[2])/32.0)*32.0 + 32.0*float64(ci)

		if w.findChunk(cx, cz) < 0 {
			w.chunkGen(cx, cz)
		}

		cx = math.Floor(float64(w.player.pos[0])/32.0)*32.0 + 32.0*float64(ci)
		cz = math.Floor(float64(w.player.pos[2])/32.0)*32.0 + 32.0*float64(ci)

		if w.findChunk(cx, cz) < 0 {
			w.chunkGen(cx, cz)
		}

		cx = math.Floor(float64(w.player.pos[0])/32.0)*32.0 - 32.0*float64(ci)
		cz = math.Floor(float64(w.player.pos[2])/32.0)*32.0 - 32.0*float64(ci)

		if w.findChunk(cx, cz) < 0 {
			w.chunkGen(cx, cz)
		}
	}
}

func (w *World) tick() {
	w.player.obj.Mesh.SphereRadius = 1.0
	w.player.obj.Mesh.SphereCenter = w.player.obj.Phys.Pos

	w.autoGen()

	hasEvil := false
	for i := 0; i < len(w.entities); i++ {
		if w.entities[i].evil {
			hasEvil = true
		}

		diff := w.player.pos.Sub(w.entities[i].obj.Phys.Pos)
		dist := diff.Len()

		if dist > 1000.0 {
			println(i, " fell")
			w.entities[i].obj.Phys.Pos[0] = float32(w.entities[i].obj.X0)
			w.entities[i].obj.Phys.Pos[1] = float32(w.entities[i].obj.Y0)
			w.entities[i].obj.Phys.Pos[2] = float32(w.entities[i].obj.Z0)
		}

		w.entities[i].obj.Phys.Rot[1] = -w.player.rot[1] + 1.57
		w.entities[i].tick()

		front := w.entities[i].obj.Phys.V.Normalize().Dot(diff.Normalize())

		if front > 0.0 {
			w.entities[i].sprite.AnimCycle = []int{0, 1, 0, 2}
		} else {
			w.entities[i].sprite.AnimCycle = []int{3, 4, 3, 5}
		}

		if w.player.attacking {
			if dist < 5.0 {
				if !w.entities[i].evil {
					w.moneyNum -= 3
				}

				var e2 grm.Obj
				//w.objs[i].LoadHull(nil, "gfx/ico.obj", "0", true, false, "")
				e2.LoadObj(nil, "gfx/soul.obj", "")
				e2.Mesh.SetCol(mgl32.Vec4{0.0, 1.5, 1.5})
				e2.Phys.RenderPos = true
				e2.SphereIsect = true
				e2.Phys.Pos = w.entities[i].obj.Phys.Pos
				e2.Phys.RPos = e2.Phys.Pos
				e2.Phys.V = mgl32.Vec3{float32(math.Sin(3.14 - float64(w.player.rot[1]))) * 0.2 * 2.0, -0.1, float32(math.Cos(3.14 - float64(w.player.rot[1]))) * 0.2 * 2.0}
				//e2.Update()

				w.souls = append(w.souls, e2)

				ranent := int(rand.Float64() * float64(len(w.entities)-1))
				newx := w.entities[ranent].obj.X0
				newy := w.entities[ranent].obj.Y0
				newz := w.entities[ranent].obj.Z0
				w.entities[ranent].obj.X0 += rand.Float64() * 64.0/*w.entities[i].obj.X0*/ // mix up coords
				w.entities[ranent].obj.Y0 += rand.Float64() * 64.0/*w.entities[i].obj.Y0*/
				w.entities[ranent].obj.Z0 += rand.Float64() * 64.0/*w.entities[i].obj.Z0*/
				w.entities[i].randomChar(newx, newy, newz)

				println("DEAD", i)

				w.sys.Clear()
				w.addPhys()

				w.soulTicks = 0.0
			}
		}
	}

	if len(w.entities) > 0 && !hasEvil {
		w.entities[0].evil = true
		w.entities[0].sprite.Mesh.SetCol(mgl32.Vec4{1.5, 1.0, 1.0, 1.0})
	}

	w.soulTicks += 0.01
	for i := 0; i < len(w.souls); i++ {
		w.souls[i].Mesh.SphereRadius = 0.1
		w.souls[i].Mesh.SphereCenter = w.souls[i].Phys.Pos
	}

	for i := 0; i < len(w.money); i++ {
		w.money[i].Phys.Rot[1] += 0.04
		w.money[i].Update()
		dist := w.player.pos.Sub(w.money[i].Phys.Pos).Len()
		if dist < 5.0 {
			bef := w.money[:i]
			aft := w.money[i+1:]
			w.money = append(bef, aft...)

			w.moneyNum += 1
		}
	}

	for i := 0; i < len(w.souls); i++ {
		w.souls[i].Phys.Rot[1] += 0.1
		dist := w.player.pos.Sub(w.souls[i].Phys.Pos).Len()
		if dist < 5.0 && w.soulTicks > 0.3 {
			bef := w.souls[:i]
			aft := w.souls[i+1:]
			w.souls = append(bef, aft...)

			w.soulsNum += 1
		}
	}

	w.sys.Update()
}

func (w *World) start() {
	w.time = 3.14
	w.soulsNum = 0
}

func (w *World) restart() {
	w.moneyNum = 10
	w.billNum = 10
	w.shiftNum = 24
	w.salaryNum = 1
	w.quotaNum = 5
	w.soulsNum = 0
	w.time = 3.14
	w.day = 0
}

func (w *World) draw(r *grm.Renderer) {
	for i := 0; i < len(w.chunks); i++ {
		w.chunks[i].Draw(r)
	}

	for i := 0; i < len(w.entities); i++ {
		w.entities[i].draw(r)

		w.entities[i].obj.InRange = w.entities[i].obj.Phys.Pos.Sub(*w.player.pos).Len() < w.far
		if !w.entities[i].obj.InRange {
			bef := w.entities[:i]
			aft := w.entities[i+1:]
			w.entities = append(bef, aft...)
		}
	}

	for i := 0; i < len(w.objs); i++ {
		w.objs[i].Draw(r)

		w.objs[i].InRange = w.objs[i].Phys.Pos.Sub(*w.player.pos).Len() < w.far
		if !w.objs[i].InRange {
			bef := w.objs[:i]
			aft := w.objs[i+1:]
			w.objs = append(bef, aft...)
		}
	}

	for i := 0; i < len(w.environs); i++ {
		w.environs[i].Draw(r)

		w.environs[i].InRange = w.environs[i].Phys.Pos.Sub(*w.player.pos).Len() < w.far
		if !w.environs[i].InRange {
			bef := w.environs[:i]
			aft := w.environs[i+1:]
			w.environs = append(bef, aft...)
		}
	}

	for i := 0; i < len(w.money); i++ {
		w.money[i].Draw(r)

		w.money[i].InRange = w.money[i].Phys.Pos.Sub(*w.player.pos).Len() < w.far
		if !w.money[i].InRange {
			bef := w.money[:i]
			aft := w.money[i+1:]
			w.money = append(bef, aft...)
		}
	}

	for i := 0; i < len(w.souls); i++ {
		w.souls[i].Draw(r)

		w.souls[i].InRange = w.souls[i].Phys.Pos.Sub(*w.player.pos).Len() < w.far
		if !w.souls[i].InRange {
			bef := w.souls[:i]
			aft := w.souls[i+1:]
			w.souls = append(bef, aft...)
		}
	}
}
