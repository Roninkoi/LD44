package main

import (
	"Gorium/game"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	g := game.Game{}

	g.Start()
}
