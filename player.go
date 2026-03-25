package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/setanarut/aseplayer"
	"github.com/setanarut/coll"
	"github.com/setanarut/kamera/v2"
	"github.com/setanarut/tween"
	"github.com/setanarut/v"
)

const (
	DashWallBlockDistance = 34
)

const (
	MoveSpeedX    = 6.125
	MoveSpeedXAir = 5.5
	MaxSpeedY     = 20.25
	JumpPower     = -11.46
	ParryPower    = -14.46
	Gravity       = 0.86
	DashSpeed     = 13.75 * 1.5
	BulletSpeed   = 13.0
	BulletRadius  = 3.0
)

// Hitbox dimensions constants
const (
	DuckHalfY = 12.0 // 24
	FallHalfY = 16.0 // 32

	// 32x70
	HalfX = 17
	HalfY = 35

	DuckTweenLerpSpeed = 0.30
)

// Durations
const (
	JumpHoldMinDuration = time.Millisecond
	JumpHoldMaxDuration = time.Millisecond * 160
	BulletDuration      = time.Millisecond * 250
	DashCooldown        = time.Millisecond * 300
	ParryDuration       = time.Millisecond * 200
	DashTimeDuration    = time.Millisecond * 300
	HitDuration         = time.Millisecond * 2000
	HitBlinkInterval    = HitDuration / 4
	Tick                = time.Second / 60
)

var bulletOffsetToggler OffsetToggler = 8.0

// Player yapısı
type Player struct {
	coll.AABB
	OldAABB     coll.AABB
	ParrySensor coll.Circle
	Hp          int

	MoveDirectionX float64
	Facing         float64
	FacingOld      float64

	Direction8 v.Vec
	Delta      v.Vec

	activeState   State
	previousState State

	fall     *fall
	grounded *grounded
	jumping  *jump
	dash     *dash
	hit      *hit

	JumpTimer           time.Duration
	DashTimer           time.Duration
	BulletTimer         time.Duration
	TimeSinceGroundDash time.Duration
	ParryTimer          time.Duration
	FireTimer           time.Duration
	HitTimer            time.Duration

	IsParrying         bool
	ParryReady         bool
	DashUsedInAir      bool
	DashEasingDisabled bool
	IsOnFloor          bool

	Paused bool

	JumpPressed     bool
	firePressed     bool
	fireJustPressed bool
	duckPressed     bool
	dashPressed     bool
	lockPressed     bool
	jumpJustPressed bool
	dashJustPressed bool

	hitInfo             coll.Hit
	groundedPlatform    any
	oldGroundedPlatform any

	dio        *ebiten.DrawImageOptions
	animPlayer *aseplayer.AnimPlayer

	currentLevel ILevel
}

func NewPlayer(pos v.Vec) *Player {
	p := &Player{
		AABB: coll.AABB{
			Pos:  pos,
			Half: v.Vec{HalfX, HalfY},
		},
		ParrySensor: coll.Circle{
			Pos:    pos,
			Radius: 43,
		},
		Facing:              1,       // 0 ile başlamasın diye
		Direction8:          v.Right, // 0 ile başlamasın diye
		TimeSinceGroundDash: time.Second * 7,
		hitInfo:             coll.Hit{},
		fall:                &fall{},
		jumping:             &jump{},
		dash:                &dash{},
		hit:                 &hit{},
		grounded: &grounded{
			activeSubState: nil,
			idle:           &idle{},
			run:            &run{},
			runTurn:        &runTurn{},
			duck:           &duck{},
			duckIdle:       &duckIdle{},
			duckTurn:       &duckTurn{},
			lock:           &lock{},
		},
		dio: &ebiten.DrawImageOptions{
			Filter: ebiten.FilterLinear,
		},
	}

	p.dash.dashTween = *tween.NewTween(0, DashSpeed, DashTimeDuration, tween.InOutSine, false)
	p.animPlayer = readAseprite("cuphead.ase")

	p.animPlayer.Play("jump")
	p.activeState = p.fall
	p.grounded.activeSubState = p.grounded.idle
	return p
}

func (p *Player) ChangeState(next State) {
	if p.activeState == next {
		return
	}
	p.previousState = p.activeState
	p.activeState.Exit(p)
	p.activeState = next
	p.activeState.Enter(p)
}

func (p *Player) IsState(s State) bool {
	return p.activeState == s
}

func (p *Player) IsSubState(s State) bool {
	return p.grounded.activeSubState == s
}

func (p *Player) fireDuck() {
	if p.BulletTimer > 0 {
		return
	}
	p.BulletTimer = BulletDuration

	bulletPos := v.Vec{p.Pos.X + (50 * p.Facing), p.Pos.Y - 3}
	bulletPos.Y += bulletOffsetToggler.Next()
	ShootManag.SpawnBullet(bulletPos, v.Vec{p.Facing * BulletSpeed, 0})
}

func (p *Player) fire8Direction() {
	if p.BulletTimer > 0 {
		return
	}
	p.BulletTimer = BulletDuration // Timer'ı SADECE burada sıfırla
	var bulletPos v.Vec
	if IsAxisHorizontal(p.Direction8) {
		bulletPos.X = p.Pos.X + (50 * p.Facing)
		bulletPos.Y = p.Pos.Y - bulletOffsetToggler.Next()
	} else if IsAxisDiagonal(p.Direction8) {
		bulletPos.X = p.Pos.X + (45 * p.Facing)
		bulletPos.Y = p.Pos.Y + (30 * p.Direction8.Y)
	} else if IsVertical(p.Direction8) {
		bulletPos = p.Pos.Add(p.Direction8.Unit().Scale(60))
		bulletPos.X += p.Facing * 15
		bulletPos.Y += bulletOffsetToggler.Next()
	}
	bulletVelocity := p.Direction8.Unit().Scale(BulletSpeed)

	if p.Delta.X != 0 {
		bulletVelocity.X += p.Delta.X
	}

	ShootManag.SpawnBullet(bulletPos, bulletVelocity)
}

func (p *Player) IsOnPlatform() bool {
	return p.groundedPlatform != nil
}
func (p *Player) StopDashing() {
	p.TimeSinceGroundDash = 0
	p.DashTimer = DashTimeDuration
}

func (p *Player) inputUpdate() {
	p.Direction8 = Axis()
	p.MoveDirectionX = getAxisX()
	p.jumpJustPressed = inpututil.IsKeyJustPressed(ebiten.KeySpace)
	p.dashJustPressed = inpututil.IsKeyJustPressed(ebiten.KeyUp)
	p.fireJustPressed = inpututil.IsKeyJustPressed(ebiten.KeyLeft)
	p.JumpPressed = ebiten.IsKeyPressed(ebiten.KeySpace)
	p.dashPressed = ebiten.IsKeyPressed(ebiten.KeyUp)
	p.lockPressed = ebiten.IsKeyPressed(ebiten.KeyDown)
	p.firePressed = ebiten.IsKeyPressed(ebiten.KeyLeft)
	p.duckPressed = ebiten.IsKeyPressed(ebiten.KeyS)
}

func (p *Player) Teleport(pos v.Vec) {
	p.Pos = pos
	p.OldAABB = p.AABB
	p.Delta = v.Vec{}
	p.IsOnFloor = false
	p.ChangeState(p.fall)
}

func (s *Player) position() v.Vec {
	return s.AABB.Pos
}

func (p *Player) IsHitTimerEnd() bool {
	return p.HitTimer <= 0
}

func (p *Player) Update() {

	if p.Paused {
		return
	}

	if p.BulletTimer > 0 {
		p.BulletTimer -= Tick
	}

	if p.HitTimer > 0 {
		p.HitTimer -= Tick
	}
	p.FacingOld = p.Facing

	p.OldAABB = p.AABB
	p.inputUpdate()

	if p.MoveDirectionX != 0 && p.activeState != p.dash {
		p.Facing = p.MoveDirectionX
	}

	if p.Direction8.IsZero() {
		p.Direction8.X = p.Facing
	}

	// Zemin dash cooldown timer'ını güncelle
	p.TimeSinceGroundDash = min(p.TimeSinceGroundDash+Tick, DashCooldown)

	// ############  FSM Durumlarını güncelle  ###################
	p.activeState.Update(p)
	p.animPlayer.Update(aseplayer.Delta)

	// Yerçekimi
	if p.activeState != p.dash {
		p.Delta.Y += Gravity
	}

	p.Delta.Y = min(p.Delta.Y, MaxSpeedY) // Speed limit
	// ############  OYUNCU HIZINI EKLE  ###################
	p.Pos = p.Pos.Add(p.Delta)

	if p.IsOnFloor && p.IsOnPlatform() {

		if shape, ok := p.groundedPlatform.(*BoxShape); ok {

			if shape.OneWay {
				if p.duckPressed && !p.lockPressed && p.jumpJustPressed {
					p.oldGroundedPlatform = p.groundedPlatform
					shape.Solid = false
				}
			}

			if !shape.parent.StaticBody {
				platformDeltaX := shape.Pos.X - shape.OldPos.X
				p.Pos.X += platformDeltaX
				// if !dp.Paused {
				p.Delta.Y = shape.Pos.Y - shape.OldPos.Y // platformun deltasını oyuncu ile eşitle
				// }
				p.SetBottom(shape.Top())
			}

		}
	}

	p.ParrySensor.Pos = p.Pos
	p.groundedPlatform = nil // reset
	p.IsOnFloor = false
}

func (p *Player) Draw(s *ebiten.Image, cam *kamera.Camera) {

	p.dio.GeoM.Reset()
	p.dio.GeoM.Translate(p.animPlayer.CurrentFrame.Position.X, p.animPlayer.CurrentFrame.Position.Y)
	p.dio.GeoM.Scale(SpriteScale*p.Facing, SpriteScale)

	switch p.activeState {
	case p.grounded:
		p.dio.GeoM.Translate(p.Pos.X, p.Bottom())
	default:
		p.dio.GeoM.Translate(p.Pos.X, p.Pos.Y)
	}

	p.dio.ColorScale.Reset()
	if p.HitTimer > 0 {
		// Her 500ms'de bir yanıp sön (4 periyot toplam)
		if (p.HitTimer/HitBlinkInterval)%2 == 0 {
			p.dio.ColorScale.ScaleAlpha(0.3)
		}
	}
	cam.Draw(p.animPlayer.CurrentFrame.Image, p.dio, s)
}
