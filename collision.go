package main

import (
	"github.com/setanarut/coll"
	"github.com/setanarut/v"
)

const PlatformTopEpsilon = 1.0

// Any platform obstacles in the facing direction?
func CheckObstacles(p *Player, dist float64, lvl ILevel) bool {
	for _, plt := range *lvl.Bodies() {
		for _, s := range plt.Shapes {
			switch shape := s.(type) {
			case *BoxShape:
				if shape.Solid {
					if !shape.HorizontalCollisionDisabled && p.groundedPlatform != plt {
						if coll.BoxBoxSweep1(&shape.AABB, &p.AABB, v.Vec{p.Facing * dist, 0}, nil) {
							return true
						}
					}
				}
			case *CircleShape:
			}
		}
	}
	return false
}

func playerCircle(cuphead *Player, circle *CircleShape, lvl ILevel) {
	// Parry Logic
	if cuphead.IsParrying && cuphead.activeState == cuphead.fall && circle.Parry {
		if coll.CircleCircleSweep2(&cuphead.ParrySensor, &circle.Circle, cuphead.Delta, circle.GetDelta(), nil) {
			cuphead.IsParrying = false
			cuphead.Delta.Y = ParryPower
			cuphead.animPlayer.Play("parry")
			GameFreeze = true

			lvl.OnParry(circle.parent)
			return
		}
	}

	// Hit Logic - Guard Clause kullanımı
	if cuphead.HitTimer > 0 {
		return
	}

	if !coll.CircleCircleSweep2(&cuphead.ParrySensor, &circle.Circle, cuphead.Delta, circle.GetDelta(), nil) {
		circle.HasPlayerEntered = false
		return
	}

	if circle.HasPlayerEntered {
		return
	}

	// Çarpışma gerçekleşti ve daha önce işlenmedi
	if circle.CanTakePlayerDamage {
		circle.parent.Hp = 0
	}
	if circle.CanGiveDamage {
		cuphead.Delta.X = 0
		cuphead.Delta.Y = -10
		cuphead.HitTimer = HitDuration
		cuphead.Hp--
		cuphead.ChangeState(cuphead.hit)
	}
	circle.HasPlayerEntered = true
}

func playerBox(cuphead *Player, box *BoxShape, l ILevel) {

	// Parry
	if cuphead.IsParrying && cuphead.activeState == cuphead.fall && box.Parry {
		if coll.BoxBoxSweep2(&box.AABB, &cuphead.AABB, box.GetDelta(), cuphead.Delta, nil) {
			cuphead.IsParrying = false
			cuphead.Delta.Y = ParryPower
			cuphead.animPlayer.Play("parry")
			GameFreeze = true
			box.parent.Hp = 0
			return
		}

	}
	if box.Solid {
		if box.CanGiveDamage {
			playerDamageBox(cuphead, box, l)
		} else {
			playerBoxDynamic(cuphead, box)
		}
	} else {
		if !box.CanGiveDamage && !box.CanTakeBulletDamage {
			playerDamageBox(cuphead, box, l)
			return
		}
		playerDamageBox(cuphead, box, l)
	}

}

func playerDamageBox(cuphead *Player, box *BoxShape, lvl ILevel) {
	if cuphead.HitTimer <= 0 {
		if coll.BoxBoxSweep2(&box.AABB, &cuphead.AABB, box.GetDelta(), cuphead.Delta, nil) {
			if !box.HasPlayerEntered {
				if box.CanTakeBulletDamage {
					box.parent.Hp = 0
				}
				if box.CanGiveDamage {
					cuphead.Delta.X = 0
					cuphead.Delta.Y = -10
					cuphead.HitTimer = HitDuration
					cuphead.Hp--
					cuphead.ChangeState(cuphead.hit)
				}
				if box.Sensor {
					lvl.OnSensorEntered(box.parent)
				}
				box.HasPlayerEntered = true
			}
		} else {
			box.HasPlayerEntered = false
		}
	}
}

func playerBoxDynamic(a *Player, bs *BoxShape) {
	aabb := &bs.AABB
	aL, aR := a.Left(), a.Right()
	bL, bR := aabb.Left(), aabb.Right()
	if aL >= bR || aR <= bL {
		return
	}
	aT, aB := a.Top(), a.Bottom()
	bT, bB := aabb.Top(), aabb.Bottom()
	if aT >= bB || aB <= bT-PlatformTopEpsilon {
		return
	}
	bDelta := bs.GetDelta()
	rVel := a.Delta.Sub(bDelta)
	if bs.OneWay {
		wasAbove := a.OldAABB.Bottom() <= (bs.OldPos.Y-aabb.Half.Y)+PlatformTopEpsilon
		isNearTop := (aB - bT) <= PlatformTopEpsilon*2
		if (wasAbove || isNearTop) && rVel.Y >= 0 {
			a.SetBottom(bT)
			a.Delta.Y = bDelta.Y
			a.IsOnFloor = true
			a.groundedPlatform = bs
		}
	} else {
		if a.OldAABB.Bottom() <= (bs.OldPos.Y-aabb.Half.Y)+PlatformTopEpsilon && rVel.Y >= 0 {
			a.SetBottom(bT)
			a.Delta.Y = bDelta.Y
			a.IsOnFloor = true
			a.groundedPlatform = bs
			return
		}
		if a.OldAABB.Top() >= (bs.OldPos.Y+aabb.Half.Y) && rVel.Y < 0 {
			a.SetTop(bB)
			a.Delta.Y = -a.Delta.Y * 0.3
			return
		}
	}
	if !bs.HorizontalCollisionDisabled {
		if a.OldAABB.Right() <= (bs.OldPos.X-aabb.Half.X) && rVel.X > 0 {
			a.SetRight(bL)
			a.StopDashing()
			return
		}
		if a.OldAABB.Left() >= (bs.OldPos.X+aabb.Half.X) && rVel.X < 0 {
			a.SetLeft(bR)
			a.StopDashing()
			return
		}
	}
	if !bs.OneWay {
		overlapT := aB - bT
		overlapB := bB - aT
		oldBottom := a.OldAABB.Bottom()
		platformOldTop := bs.OldPos.Y - aabb.Half.Y
		if bs.HorizontalCollisionDisabled {
			if overlapT <= overlapB && rVel.Y >= 0 && oldBottom <= platformOldTop+PlatformTopEpsilon {
				a.SetBottom(bT)
				a.IsOnFloor = true
				a.groundedPlatform = bs
			} else if overlapB < overlapT && rVel.Y <= 0 {
				a.SetTop(bB)
				a.Delta.Y = -a.Delta.Y * 0.3
			}
		} else {
			overlapL := aR - bL
			overlapR := bR - aL
			minOverlap := min(min(overlapT, overlapB), min(overlapL, overlapR))
			if minOverlap == overlapT && rVel.Y >= 0 && oldBottom <= platformOldTop+PlatformTopEpsilon {
				a.SetBottom(bT)
				a.IsOnFloor = true
				a.groundedPlatform = bs
			} else if minOverlap == overlapB && rVel.Y <= 0 {
				a.SetTop(bB)
				a.Delta.Y = -a.Delta.Y * 0.3
			} else if minOverlap == overlapL && rVel.X >= 0 {
				a.SetRight(bL)
				a.StopDashing()
			} else if minOverlap == overlapR && rVel.X <= 0 {
				a.SetLeft(bR)
				a.StopDashing()
			}
		}
	}
}
