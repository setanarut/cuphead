package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/setanarut/v"
)

var timerR int
var timerL int

func getAxisX() float64 {
	x := 0.0
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		timerR++
		x = 1.0
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		timerL++
		x = -1.0
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyD) {
		timerR = 0
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyA) {
		timerL = 0
	}

	if ebiten.IsKeyPressed(ebiten.KeyA) && ebiten.IsKeyPressed(ebiten.KeyD) {
		if timerL > timerR {
			x = -1.0
		}
		if timerR > timerL {
			x = 1.0
		}
	}
	return x
}

// var lastAxis v.Vec

func Axis() (axis v.Vec) {
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		axis.Y -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		axis.Y += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		axis.X -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		axis.X += 1
	}
	return
}

func IsAxisHorizontal(axis v.Vec) bool {
	return axis.X != 0 && axis.Y == 0
}

func IsAxisDiagonal(axis v.Vec) bool {
	return axis.X != 0 && axis.Y != 0
}
func IsVertical(axis v.Vec) bool {
	return axis.X == 0 && axis.Y != 0
}
