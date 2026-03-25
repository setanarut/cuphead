package main

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/setanarut/aseplayer"
	"github.com/setanarut/coll"
	"github.com/setanarut/tween"
	"github.com/setanarut/v"
)

type ShapeType uint8

type IShape interface {
	GetParent() *Body
	GetShapeData() *ShapeData
	Update()
}

type Body struct {
	Tag              string
	Pos              v.Vec
	Hp               int
	ImgIndex         int
	ImgScale         float64
	ImgOffset        v.Vec
	DamageFlashTimer time.Duration
	Shapes           []IShape
	StaticBody       bool
	StaticImage      bool
	Mover            Mover
	Anim             aseplayer.AnimPlayer
}

func (b *Body) ResetSpriteOffsets() {
	b.ImgOffset = v.Vec{}
	b.ImgScale = 1
}
func (b *Body) Update() {

	if !b.StaticBody && b.Mover != nil {
		b.Pos = b.Mover.Move(b.Pos)
	}

	if b.DamageFlashTimer > 0 {
		b.DamageFlashTimer -= Tick
	}

	for _, shape := range b.Shapes {
		shape.Update()
	}

}

func (b *Body) AddShape(s IShape) {
	b.Shapes = append(b.Shapes, s)
}
func (b *Body) DamageFlash() {
	b.DamageFlashTimer = DamageFlashDuration
}

func NewBody(pos v.Vec, hp int, static bool, shapes ...IShape) *Body {
	b := &Body{
		Pos:      pos,
		Hp:       hp,
		Shapes:   shapes,
		ImgIndex: -1,
		ImgScale: 1,
	}

	for _, v := range b.Shapes {
		v.GetShapeData().parent = b
	}

	if !static {
		b.Mover = &SegTweenMover{
			Seg: coll.Segment{pos, v.Vec{pos.X, pos.Y + 100}},
			Tw:  *tween.NewTween(0, 1, time.Millisecond*1500, tween.InOutCubic, true),
		}
	} else {
		b.StaticBody = true
	}
	return b
}

type ShapeData struct {
	parent                      *Body
	OldPos                      v.Vec
	LocalOffset                 v.Vec
	HasPlayerEntered            bool
	Solid                       bool
	Sensor                      bool
	OneWay                      bool
	HorizontalCollisionDisabled bool
	CanTakeBulletDamage         bool
	CanTakePlayerDamage         bool
	CanGiveDamage               bool
	Parry                       bool
	IgnoreBullet                bool
}

func (s *ShapeData) GetParent() *Body { return s.parent }

type BoxShape struct {
	coll.AABB
	*ShapeData
}

// BoxShape Interface Uygulamaları
func (b *BoxShape) SetPos(p v.Vec)           { b.Pos = p }
func (b *BoxShape) GetShapeData() *ShapeData { return b.ShapeData }
func (b *BoxShape) GetDelta() v.Vec {
	return b.Pos.Sub(b.OldPos)
}
func (b *BoxShape) Update() {
	b.OldPos = b.Pos
	b.Pos = b.parent.Pos.Add(b.LocalOffset).Floor()
}

func NewBoxShape(parent *Body, halfSize v.Vec, solid bool) *BoxShape {
	bs := &BoxShape{
		ShapeData: &ShapeData{
			parent: parent,
			// Type:          TypeBox,
			Solid:               solid,
			CanTakeBulletDamage: true,
			OldPos:              parent.Pos,
		},
		AABB: coll.AABB{
			Pos:  parent.Pos,
			Half: halfSize,
		},
	}

	return bs
}

type CircleShape struct {
	*ShapeData
	coll.Circle
}

func (b *CircleShape) GetShapeData() *ShapeData { return b.ShapeData }
func (b *CircleShape) GetDelta() v.Vec {
	return b.Pos.Sub(b.OldPos)
}
func (b *CircleShape) Update() {
	b.OldPos = b.Pos
	b.Pos = b.parent.Pos.Add(b.LocalOffset)
}

func NewCircleShape(parent *Body, radius float64, solid bool, offset v.Vec) *CircleShape {
	return &CircleShape{
		ShapeData: &ShapeData{
			LocalOffset: offset,
			parent:      parent,
			// Type:        TypeCircle, // Buradaki hatayı düzelttik (TypeBox yazıyordu)
			Solid: solid,
		},
		Circle: coll.Circle{
			Pos:    parent.Pos,
			Radius: radius,
		},
	}
}

func (b *Body) Clone() *Body {
	var buf bytes.Buffer
	var cloneBody Body

	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	_ = enc.Encode(b)
	_ = dec.Decode(&cloneBody)

	for _, s := range cloneBody.Shapes {
		s.GetShapeData().parent = &cloneBody
	}

	switch mover := b.Mover.(type) {
	case *SegTweenMover:
		cloneMover := *mover
		cloneBody.Mover = &cloneMover
	case *OrbitalMover:
		cloneMover := *mover
		cloneBody.Mover = &cloneMover
	case *PathMover:
		cloneMover := *mover
		cloneBody.Mover = &cloneMover
	case nil:
		cloneBody.Mover = nil
	}

	return &cloneBody
}

func (b *Body) Clone2() *Body {
	newBody := *b
	newBody.Shapes = make([]IShape, len(b.Shapes))

	switch mover := b.Mover.(type) {
	case *SegTweenMover:
		cloneMover := *mover
		newBody.Mover = &cloneMover
	case *OrbitalMover:
		cloneMover := *mover
		newBody.Mover = &cloneMover
	case *PathMover:
		cloneMover := *mover
		newBody.Mover = &cloneMover
	case nil:
		newBody.Mover = nil
	}

	for i, s := range b.Shapes {
		switch shape := s.(type) {
		case *BoxShape:
			cloneShape := *shape
			cloneShapeData := *shape.ShapeData
			cloneShapeData.parent = &newBody
			cloneShape.ShapeData = &cloneShapeData
			newBody.Shapes[i] = &cloneShape
		case *CircleShape:
			nc := *shape
			nd := *shape.ShapeData
			nd.parent = &newBody
			nc.ShapeData = &nd
			newBody.Shapes[i] = &nc
		}
	}
	return &newBody
}
