package main

import (
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/setanarut/v"
)

type Sprite struct {
	Pos v.Vec
	Img *ebiten.Image
}

func loadSprite(p v.Vec, f string) Sprite {
	im, _, _ := ebitenutil.NewImageFromFileSystem(assets, f)
	return Sprite{
		Pos: p,
		Img: im,
	}
}
