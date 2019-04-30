package grm

import (
	"github.com/go-gl/mathgl/mgl32"
)

const texW = 2048
const texH = 2048

type Sprite struct {
	Mesh Mesh

	anim        int
	animTime    float64
	timeOld     float64
	AnimCycle   []int
	animSprites []mgl32.Vec4
}

func (s *Sprite) animTick() {
	s.anim += 1

	if s.anim >= len(s.AnimCycle) {
		s.anim = 0
	}
}

func (s *Sprite) animGetTex() {
	x := s.animSprites[s.AnimCycle[s.anim]][0] / (float32)(texW)
	y := -s.animSprites[s.AnimCycle[s.anim]][1] / (float32)(texH)
	w := s.animSprites[s.AnimCycle[s.anim]][2] / (float32)(texW)
	h := s.animSprites[s.AnimCycle[s.anim]][3] / (float32)(texH)

	t0 := mgl32.Vec4{x / w, y / h, w, h}
	t1 := mgl32.Vec4{x/w + 1.0, y / h, w, h}
	t2 := mgl32.Vec4{x/w + 1.0, y/h + 1.0, w, h}
	t3 := mgl32.Vec4{x / w, y/h + 1.0, w, h}

	s.Mesh.texData = []float32{
		t1[0], t1[1], t1[2], t1[3],
		t3[0], t3[1], t3[2], t3[3],
		t2[0], t2[1], t2[2], t2[3],
		t0[0], t0[1], t0[2], t0[3],
	}
}

func (s *Sprite) AnimUpdate() {
	if timeNow()-s.timeOld >= s.animTime {
		s.animTick()
		s.timeOld = timeNow()
	}

	s.animGetTex()
}

func (s *Sprite) AnimLoad(ac []int, at float64, as []mgl32.Vec4) {
	s.AnimCycle = ac

	s.animTime = at

	s.animSprites = as

	s.animGetTex()
}

func (s *Sprite) AnimDraw(r *Renderer) {
	s.Mesh.Draw(r)
}

func (s *Sprite) LoadSprite(r *Renderer, t string) {
	s.Mesh.LoadMesh(r, "gfx/quad.obj", "")
}

func (s *Sprite) LoadTextSprite(r *Renderer, t string) {
	s.Mesh.LoadMesh(r, "gfx/textquad.obj", "")
}
