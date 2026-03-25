package main

import (
	"time"

	"github.com/setanarut/tween"
	"github.com/setanarut/v"
)

var (
	DownRight = v.Vec{1, 1}
	DownLeft  = v.Vec{-1, 1}
	UpRight   = v.Vec{1, -1}
	UpLeft    = v.Vec{-1, -1}
)

// State Interface
type State interface {
	Enter(p *Player)
	Update(p *Player)
	Exit(p *Player)
}

type fall struct {
}

func (s *fall) Update(p *Player) {

	if p.IsParrying {
		p.animPlayer.PlayIfNotCurrent("parry")
	} else {
		p.animPlayer.PlayIfNotCurrent("jump")
	}

	// Movement
	p.Delta.X = p.MoveDirectionX * MoveSpeedXAir

	// Zemine indik mi?
	if p.IsOnFloor {
		p.ChangeState(p.grounded)
		return
	}

	if p.dashJustPressed && !p.DashUsedInAir && p.TimeSinceGroundDash >= DashCooldown {
		if !CheckObstacles(p, DashWallBlockDistance, p.currentLevel) {
			p.ChangeState(p.dash)
		}
	}

	if p.jumpJustPressed && !p.IsParrying && p.ParryReady {
		p.IsParrying = true
	}

	if p.IsParrying {
		p.ParryTimer += Tick

	}

	// Parry ends
	if p.ParryTimer >= ParryDuration {

		p.ParryTimer = 0
		p.IsParrying = false
		p.ParryReady = false
	}

	if p.firePressed {
		p.fire8Direction()
	}

}

func (s *fall) Enter(p *Player) {
	p.animPlayer.Play("jump")
	if p.previousState != p.jumping {
		// resizeBoxFixedBottom(p.AABB, FallHalfY)
		// resizeBoxFixedTop(p.AABB, FallHalfY)
		p.Half.Y = FallHalfY
	}
	p.ParryReady = true
}
func (s *fall) Exit(p *Player) {
	p.ParryTimer = 0
	p.IsParrying = false
	p.ParryReady = false
}

type jump struct{}

func (s *jump) Update(p *Player) {

	// Horizontal air Movement
	p.Delta.X = p.MoveDirectionX * MoveSpeedX
	p.Delta.Y = JumpPower

	// Timer
	p.JumpTimer += Tick

	// Transitions
	if p.dashJustPressed && !p.DashUsedInAir {
		if !CheckObstacles(p, DashWallBlockDistance, p.currentLevel) {
			p.ChangeState(p.dash)
		}
		return
	}

	release := (p.JumpTimer >= JumpHoldMinDuration && !p.JumpPressed) ||
		(p.JumpTimer >= JumpHoldMaxDuration)
	if release {
		p.ChangeState(p.fall)
	}

}

func (s *jump) Enter(p *Player) {
	p.animPlayer.Play("jump")
	p.AABB.Half.Y = FallHalfY
	p.Delta.Y = JumpPower
	p.IsOnFloor = false
	p.JumpTimer = 0
	p.groundedPlatform = nil
}

func (s *jump) Exit(p *Player) {
	p.JumpTimer = 0
}

type hit struct {
}

func (s *hit) Enter(p *Player) {
	p.animPlayer.Play("hit")
}
func (s *hit) Update(p *Player) {

	if p.animPlayer.IsEnded() {
		if p.IsOnFloor {
			p.ChangeState(p.grounded)
		} else {
			p.ChangeState(p.fall)
		}
	}
}
func (s *hit) Exit(p *Player) {
}

type dash struct {
	dashDirectionX float64
	dashTween      tween.Tween
}

func (s *dash) Enter(p *Player) {
	s.dashTween.Reset()
	p.animPlayer.Play("dash")
	s.dashDirectionX = p.Facing
	resizeBoxFixedBottom(&p.AABB, DuckHalfY)
	p.Delta.Y = 0
	if !p.IsOnFloor {
		p.DashUsedInAir = true
	}

}

func (s *dash) Update(p *Player) {
	s.dashTween.Update(Tick)
	p.Delta.X = s.dashDirectionX * s.dashTween.Value
	p.Delta.Y = 0
	if s.dashTween.IsFinished() {
		if !p.IsOnPlatform() {
			p.ChangeState(p.fall)
		} else {
			p.ChangeState(p.grounded)
		}
	}
}

func (s *dash) Exit(p *Player) {
	p.TimeSinceGroundDash = 0
	p.Delta.X = 0

}

type grounded struct {
	previousSubState, activeSubState  State
	currentDuckHalfY, targetDuckHalfY float64
	// Alt durumlar
	*idle
	*run
	*runTurn
	*duck
	*duckIdle
	*duckTurn
	*lock
}

func (g *grounded) ChangeSubState(p *Player, next State) {
	g.previousSubState = g.activeSubState
	g.activeSubState.Exit(p)
	g.activeSubState = next
	g.activeSubState.Enter(p)
}
func (g *grounded) IsActiveState(st State) bool {
	return g.activeSubState == st
}

func (g *grounded) Update(p *Player) {

	if p.dashJustPressed && p.TimeSinceGroundDash >= DashCooldown {
		if p.grounded.activeSubState != p.grounded.duck {
			if !CheckObstacles(p, DashWallBlockDistance, p.currentLevel) {
				p.ChangeState(p.dash)
				return
			}
		}
	}

	// Zemin kontrolü
	if !p.IsOnFloor {
		p.ChangeState(p.fall)
		return
	}
	// Zemin kontrolü
	if p.jumpJustPressed && !p.lockPressed && !p.duckPressed {
		p.ChangeState(p.jumping)
		return
	}

	// Alt durumu çalıştır
	g.activeSubState.Update(p)

}
func (g *grounded) Enter(p *Player) {

	if p.TimeSinceGroundDash >= DashCooldown {
		p.TimeSinceGroundDash = time.Second
	}
	if p.oldGroundedPlatform != nil {
		if plat, ok := p.oldGroundedPlatform.(*BoxShape); ok {
			plat.Solid = true
		}
	}

	if !p.duckPressed {
		resizeBoxFixedBottom(&p.AABB, 35)
	}

	p.DashUsedInAir = false // Havadaki dash hakkı yenilendi
	p.JumpTimer = 0

	g.activeSubState.Enter(p)
}

func (g *grounded) Exit(p *Player) {
	g.activeSubState.Exit(p)
}

// IDLE (Başlangıç ve Geçiş Durumu)
type idle struct{}

func (s *idle) Enter(p *Player) {
	resizeBoxFixedBottom(&p.AABB, HalfY)
	p.animPlayer.Play("idle")
}
func (s *idle) Update(p *Player) {
	if p.firePressed {
		switch p.Direction8 {
		case v.Right, v.Left:
			p.animPlayer.PlayIfNotCurrent("shoot_straight")
		case v.Up:
			p.animPlayer.PlayIfNotCurrent("shoot_up")
		}
	} else {
		p.animPlayer.PlayIfNotCurrent("idle")
	}

	p.Delta.X = 0
	if p.MoveDirectionX != 0 {
		p.grounded.ChangeSubState(p, p.grounded.run)
	}
	if p.duckPressed {
		p.grounded.ChangeSubState(p, p.grounded.duck)
	}
	if p.lockPressed {
		p.grounded.ChangeSubState(p, p.grounded.lock)
	}

	if p.firePressed {
		p.fire8Direction()
	}

}
func (s *idle) Exit(p *Player) {}

// RUN
type run struct{}

func (s *run) Enter(p *Player) {
	p.animPlayer.Play("run")
}

func (s *run) Update(p *Player) {

	if p.firePressed {
		switch p.Direction8 {
		case UpRight, UpLeft:
			p.animPlayer.PlayIfNotCurrent("run_shoot_diag_up")
		default:
			p.animPlayer.PlayIfNotCurrent("run_shoot")
		}
	} else {
		p.animPlayer.PlayIfNotCurrent("run")
	}

	if p.lockPressed {
		p.grounded.ChangeSubState(p, p.grounded.lock)
		return
	}

	p.Delta.X = p.MoveDirectionX * MoveSpeedX

	if p.Facing != p.FacingOld {
		p.grounded.ChangeSubState(p, p.grounded.runTurn)
	}

	if p.MoveDirectionX == 0 {
		p.grounded.ChangeSubState(p, p.grounded.idle)
	}
	if p.duckPressed {
		p.grounded.ChangeSubState(p, p.grounded.duck)
	}

	if p.firePressed {
		p.fire8Direction()
	}

	// // Jump
	// if p.jumpJustPressed {
	// 	p.ChangeState(p.jumping)
	// }
}

func (s *run) Exit(p *Player) {}

// RUN TURN
type runTurn struct{}

func (s *runTurn) Enter(p *Player) {
	if p.firePressed {
		switch p.Direction8 {
		case v.Right, v.Left:
			p.animPlayer.Play("run_shoot_turn")
		case UpRight, UpLeft:
			p.animPlayer.Play("run_shoot_diag_turn")
		}
	} else {
		p.animPlayer.Play("run_turn")
	}
}

func (s *runTurn) Update(p *Player) {
	if p.animPlayer.IsEnded() {
		p.grounded.ChangeSubState(p, p.grounded.run)
	}
	if p.firePressed {
		p.fire8Direction()
	}
}

func (s *runTurn) Exit(p *Player) {}

// DUCK STATE
type duck struct {
	t float64
}

func (s *duck) Enter(p *Player) {
	s.t = 16
	p.Delta.X = 0
	resizeBoxFixedBottom(&p.AABB, 20)
	p.animPlayer.Play("duck")
}

func (s *duck) Update(p *Player) {

	s.t = max(s.t-1.1, DuckHalfY)
	resizeBoxFixedBottom(&p.AABB, s.t)
	if p.animPlayer.IsEnded() || !p.duckPressed {
		p.grounded.ChangeSubState(p, p.grounded.duckIdle)
	}

	if p.firePressed {
		p.fireDuck()
	}

}

func (s *duck) Exit(p *Player) {

}

// DUCK IDLE
type duckIdle struct {
}

func (s *duckIdle) Enter(p *Player) {
	resizeBoxFixedBottom(&p.AABB, DuckHalfY)
	p.animPlayer.Play("duck_idle")
}
func (s *duckIdle) Update(p *Player) {

	if p.firePressed {
		p.fireDuck()
		p.animPlayer.PlayIfNotCurrent("duck_shoot")
	} else {
		p.animPlayer.PlayIfNotCurrent("duck_idle")
	}

	if !p.duckPressed {
		p.grounded.ChangeSubState(p, p.grounded.idle)
	}

	if p.Facing != p.FacingOld {
		p.grounded.ChangeSubState(p, p.grounded.duckTurn)
	}

}

func (s *duckIdle) Exit(p *Player) {

}

// DUCK TURN
type duckTurn struct{}

func (s *duckTurn) Enter(p *Player) {
	p.animPlayer.Play("duck_turn")
}
func (s *duckTurn) Update(p *Player) {

	if p.animPlayer.IsEnded() {
		p.grounded.ChangeSubState(p, p.grounded.duckIdle)
	}

	if p.firePressed {
		p.fireDuck()
		p.grounded.ChangeSubState(p, p.grounded.duckIdle)

	}

}
func (s *duckTurn) Exit(p *Player) {}

// LOCK (Aim)
type lock struct{}

func (s *lock) Enter(p *Player) {
	p.Delta.X = 0
}

func (s *lock) Update(p *Player) {
	if !p.lockPressed {
		p.grounded.ChangeSubState(p, p.grounded.idle)
		return
	}

	if p.firePressed {

		switch p.Direction8 {
		case v.Right, v.Left:
			p.animPlayer.PlayIfNotCurrent("shoot_straight")
		case UpRight, UpLeft:
			p.animPlayer.PlayIfNotCurrent("shoot_diag_up")
		case DownRight, DownLeft:
			p.animPlayer.PlayIfNotCurrent("shoot_diag_down")
		case v.Down:
			p.animPlayer.PlayIfNotCurrent("shoot_down")
		case v.Up:
			p.animPlayer.PlayIfNotCurrent("shoot_up")
		}

		p.fire8Direction()

	} else {
		switch p.Direction8 {
		case v.Right, v.Left:
			p.animPlayer.PlayIfNotCurrent("aim_straight")
		case UpRight, UpLeft:
			p.animPlayer.PlayIfNotCurrent("aim_diag_up")
		case DownRight, DownLeft:
			p.animPlayer.PlayIfNotCurrent("aim_diag_down")
		case v.Down:
			p.animPlayer.PlayIfNotCurrent("aim_down")
		case v.Up:
			p.animPlayer.PlayIfNotCurrent("aim_up")
		}
	}

}
func (s *lock) Exit(p *Player) {}
