package main

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/setanarut/aseplayer"
	"github.com/setanarut/coll"
	"github.com/setanarut/kamera/v2"
	"github.com/setanarut/v"
)

type ShootManager struct {
	Bullets            []*Bullet
	BulletSpawnEffects []*BulletSpawnEffect
	BulletDeathEffects []*BulletDeathEffect

	Anim *aseplayer.AnimPlayer
	dio  *ebiten.DrawImageOptions
}

func NewShootManager() *ShootManager {
	sm := &ShootManager{
		Bullets:            make([]*Bullet, 0, 32),
		BulletSpawnEffects: make([]*BulletSpawnEffect, 0, 32),
		BulletDeathEffects: make([]*BulletDeathEffect, 0, 32),
		dio:                &ebiten.DrawImageOptions{},
	}

	sm.Anim, _ = aseplayer.NewAnimPlayerFromAsepriteFileSystem(assets, "img/bullet.ase")
	parseUserDataOffsets(sm.Anim)

	return sm
}

func (s *ShootManager) Update() {

	for _, b := range s.Bullets {
		b.Update()
	}
	for _, b := range s.BulletSpawnEffects {
		b.Update(Cuphead)
	}
	for _, b := range s.BulletDeathEffects {
		b.Update()
	}

	s.BulletSpawnEffects = slices.DeleteFunc(
		s.BulletSpawnEffects,
		func(b *BulletSpawnEffect) bool {
			return b.anim.IsEnded()
		},
	)

	s.BulletDeathEffects = slices.DeleteFunc(
		s.BulletDeathEffects,
		func(b *BulletDeathEffect) bool {
			return b.anim.IsEnded()
		},
	)

}
func (s *ShootManager) Draw(screen *ebiten.Image, cam *kamera.Camera) {

	// Spawn efektlerini çiz
	for _, b := range s.BulletSpawnEffects {
		pivot := b.anim.CurrentFrame.Position
		s.dio.GeoM.Reset()
		s.dio.GeoM.Translate(pivot.X, pivot.Y)
		s.dio.GeoM.Scale(SpriteScale, SpriteScale)
		s.dio.GeoM.Translate(b.Pos.X, b.Pos.Y)

		if b.anim.CurrentFrame.Image != nil {
			cam.Draw(b.anim.CurrentFrame.Image, s.dio, screen)
		}
	}

	for _, b := range s.BulletDeathEffects {
		pivot := b.anim.CurrentFrame.Position
		s.dio.GeoM.Reset()
		s.dio.GeoM.Translate(pivot.X, pivot.Y)
		s.dio.GeoM.Scale(SpriteScale, SpriteScale)
		s.dio.GeoM.Translate(b.Pos.X, b.Pos.Y)

		if b.anim.CurrentFrame.Image != nil {
			cam.Draw(b.anim.CurrentFrame.Image, s.dio, screen)
		}
	}

	for _, b := range s.Bullets {
		pivot := b.anim.CurrentFrame.Position
		s.dio.GeoM.Reset()
		s.dio.GeoM.Translate(pivot.X, pivot.Y)
		s.dio.GeoM.Scale(SpriteScale, SpriteScale)
		s.dio.GeoM.Rotate(b.vel.Angle())
		s.dio.GeoM.Translate(b.Pos.X, b.Pos.Y)

		if b.anim.CurrentFrame.Image != nil {
			cam.Draw(b.anim.CurrentFrame.Image, s.dio, screen)
		}
	}

}

func (s *ShootManager) SpawnBullet(pos, vel v.Vec) {

	bullet := &Bullet{
		Circle: coll.Circle{pos, BulletRadius},
		vel:    vel,
		anim:   *s.Anim,
	}

	spawnEffect := &BulletSpawnEffect{
		Pos:  pos,
		anim: *s.Anim,
	}

	bullet.anim.Play("bullet")
	spawnEffect.anim.Play("bullet_spawn")

	s.Bullets = append(s.Bullets, bullet)
	s.BulletSpawnEffects = append(s.BulletSpawnEffects, spawnEffect)
}

func (s *ShootManager) SpawnDeathEffect(pos v.Vec) {
	death := &BulletDeathEffect{
		Pos:  pos,
		anim: *s.Anim,
	}
	death.anim.Play("bullet_death")
	s.BulletDeathEffects = append(s.BulletDeathEffects, death)
}

type Bullet struct {
	coll.Circle
	vel  v.Vec
	anim aseplayer.AnimPlayer
}

func (b *Bullet) Update() {
	b.Pos = b.Pos.Add(b.vel)
	b.anim.Update(aseplayer.Delta)
	if b.anim.IsEnded() {
		b.anim.PlayIfNotCurrent("bullet_travel")
	}
}

type BulletSpawnEffect struct {
	Pos  v.Vec
	anim aseplayer.AnimPlayer
}

func (b *BulletSpawnEffect) Update(p *Player) {

	if p.activeState == p.grounded {
		b.Pos = b.Pos.Add(p.Delta)
		// if dp, ok := p.groundedPlatform.(*DynamicPlatform); ok {
		// b.Pos = b.Pos.Add(dp.Pos.Sub(dp.OldPos))
		// }
	}

	b.anim.Update(aseplayer.Delta)
}

type BulletDeathEffect struct {
	Pos  v.Vec
	anim aseplayer.AnimPlayer
}

func (b *BulletDeathEffect) Update() {
	b.anim.Update(aseplayer.Delta)
}
