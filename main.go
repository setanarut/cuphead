package main

import (
	"embed"
	"encoding/gob"
	"image/color"
	_ "image/png"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/setanarut/kamera/v2"
	"github.com/setanarut/v"
)

const DamageOpacity = 0.3
const DamageOpacityScale = 1.0 - DamageOpacity
const SpriteScale = 0.7

//go:embed img
var assets embed.FS

var (
	Cuphead      *Player
	MainCamera   *kamera.Camera
	CameraTarget v.Vec
	ShootManag   *ShootManager

	TutorialLvl *TutorialLevel
)

var (
	GamePaused      bool
	GameFreeze      bool
	GameFreezeTimer time.Duration
	screenSize          = v.Vec{X: 1024, Y: 600}
	screenFillColor int = 240
)

type Game struct {
	levelEditor       *LevelEditor
	opt               *ebiten.RunGameOptions
	enableLevelEditor bool

	currentLevel ILevel
}

func init() {
	gob.Register(&BoxShape{})
	gob.Register(&CircleShape{})
	gob.Register(&SegTweenMover{})
	gob.Register(&PathMover{})
	gob.Register(&OrbitalMover{})
	gob.Register(&TutorialLevel{})

	MainCamera = kamera.NewCamera(0, 0, screenSize.X, screenSize.Y)
	ShootManag = NewShootManager()
	Cuphead = NewPlayer(v.Vec{})

	TutorialLvl = &TutorialLevel{}
	// TutorialLvl.Load()
	TutorialLvl.Init()

	Cuphead.currentLevel = TutorialLvl

}

func (g *Game) Init() {

	g.enableLevelEditor = true
	g.opt = &ebiten.RunGameOptions{GraphicsLibrary: ebiten.GraphicsLibraryMetal}

	g.levelEditor = &LevelEditor{}
	g.levelEditor.Init(TutorialLvl)

	g.currentLevel = TutorialLvl

}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Gray{uint8(screenFillColor)})
	g.currentLevel.Draw(screen)
	Cuphead.Draw(screen, MainCamera)
	ShootManag.Draw(screen, MainCamera)
	g.levelEditor.Draw(screen, g.currentLevel)

}

func (g *Game) Update() error {

	g.levelEditor.Update(TutorialLvl)

	UpdateILevel(g.currentLevel)

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return int(screenSize.X), int(screenSize.Y)
}

func (g *Game) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	return screenSize.X, screenSize.Y
}

func main() {
	game := &Game{}
	game.Init()
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetWindowSize(int(screenSize.X), int(screenSize.Y))
	ebiten.SetRunnableOnUnfocused(false)
	if err := ebiten.RunGameWithOptions(game, game.opt); err != nil {
		log.Fatal(err)
	}
}
