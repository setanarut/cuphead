package main

import (
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/setanarut/coll"
)

const (
	ParryFreezeDuration = time.Millisecond * 150
	DamageFlashDuration = time.Millisecond * 100
)

type ILevel interface {
	Init()
	Update()
	Draw(s *ebiten.Image)
	Bodies() *[]*Body
	OnBodyRemove(b *Body)
	OnParry(b *Body)
	OnDamage(b *Body)
	OnSensorEntered(b *Body)
	Save()
	Load()
}

func UpdateILevel(lvl ILevel) {

	if GamePaused {
		return
	}

	if GameFreeze {
		GameFreezeTimer += Tick
		if GameFreezeTimer >= ParryFreezeDuration {
			GameFreeze = false
			GameFreezeTimer = 0
		}
	}
	if !GameFreeze {

		for _, b := range *lvl.Bodies() {
			b.Update()
		}
		lvl.Update()

		Cuphead.Update()
		ShootManag.Update()

		for _, body := range *lvl.Bodies() {
			for _, sh := range body.Shapes {
				switch shape := sh.(type) {
				// level.go içindeki switch bloğu yerine:
				case *BoxShape:
					playerBox(Cuphead, shape, lvl)

					//  PLAYER-BOX
					if !shape.IgnoreBullet {
						ShootManag.Bullets = slices.DeleteFunc(ShootManag.Bullets, func(blt *Bullet) bool {
							// 1. Ekran dışı kontrolü (Önce ucuz işlem)
							if blt.Pos.X > MainCamera.Right() || blt.Pos.X < MainCamera.X ||
								blt.Pos.Y < MainCamera.Y || blt.Pos.Y > MainCamera.Bottom() {
								ShootManag.SpawnDeathEffect(blt.Pos)
								return true
							}

							// 2. Çarpışma yoksa devam et
							if !shape.Solid || !coll.BoxCircleSweep2(&shape.AABB, &blt.Circle, shape.GetDelta(), blt.vel, nil) {
								return false
							}

							// 3. Çarpışma gerçekleşti
							if shape.CanTakeBulletDamage {
								body.Hp--
								body.DamageFlash()
							}
							ShootManag.SpawnDeathEffect(blt.Pos)
							return true
						})
					}
				case *CircleShape:

					//  PLAYER-CIRCLE
					playerCircle(Cuphead, shape, lvl)

					//  BULLETS-CIRCLE
					ShootManag.Bullets = slices.DeleteFunc(
						ShootManag.Bullets,
						func(blt *Bullet) bool {
							if !shape.IgnoreBullet {
								if coll.CircleCircleSweep2(&shape.Circle, &blt.Circle, shape.GetDelta(), blt.vel, nil) {
									if shape.Solid {
										if shape.CanTakeBulletDamage {
											body.Hp--
											body.DamageFlash()
											lvl.OnDamage(body)
										}
										ShootManag.SpawnDeathEffect(blt.Pos)
										return true
									}
								}
							}
							offScreen := blt.Pos.X > MainCamera.Right() ||
								blt.Pos.X < MainCamera.X ||
								blt.Pos.Y < MainCamera.Y ||
								blt.Pos.Y > MainCamera.Bottom()
							if offScreen {
								ShootManag.SpawnDeathEffect(blt.Pos)
							}
							return offScreen
						},
					)

				}
			}
		}

		// Remove bodies
		bd := lvl.Bodies()

		*bd = slices.DeleteFunc(*bd, func(body *Body) bool {
			shouldRemove := body.Hp <= 0
			if shouldRemove {
				lvl.OnBodyRemove(body)
			}
			return shouldRemove
		})

	}

}
