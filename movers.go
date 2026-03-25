package main

import (
	"math"

	"github.com/setanarut/coll"
	"github.com/setanarut/gog/v2/path"
	"github.com/setanarut/tween"
	"github.com/setanarut/v"
)

type Mover interface {
	Move(pos v.Vec) (newPos v.Vec)
	SetPos(v.Vec)
	Pos() v.Vec
}

type OrbitalMover struct {
	Orbit *coll.Circle
	Speed float64
	Angle float64
}

func (m *OrbitalMover) Move(pos v.Vec) v.Vec {
	m.Angle += m.Speed
	pos.X = m.Orbit.Pos.X + math.Cos(m.Angle)*m.Orbit.Radius
	pos.Y = m.Orbit.Pos.Y + math.Sin(m.Angle)*m.Orbit.Radius
	return pos
}

func (m *OrbitalMover) SetPos(pos v.Vec) {
	m.Orbit.Pos = pos
}
func (m *OrbitalMover) Pos() v.Vec {
	return m.Orbit.Pos
}

type PathMover struct {
	Path  *path.Path
	Index int
}

func (m *PathMover) SetPos(pos v.Vec) {
	m.Path.SetPos(pos)
}
func (m *PathMover) Pos() v.Vec {
	return m.Path.Anchor
}

func (m *PathMover) Move(pos v.Vec) v.Vec {
	currentPoint := m.Path.Points[m.Index]
	m.Index = (m.Index + 1) % len(m.Path.Points)
	return currentPoint
}

type SegTweenMover struct {
	Seg    coll.Segment
	Tw     tween.Tween
	Paused bool
}

func (m *SegTweenMover) Move(pos v.Vec) v.Vec {
	if m.Paused {
		return pos
	}
	m.Tw.Update(Tick)
	return m.Seg.A.Lerp(m.Seg.B, m.Tw.Value)
}

func (m *SegTweenMover) SetPos(pos v.Vec) {
	offset := pos.Sub(m.Seg.A)
	m.Seg.A = pos
	m.Seg.B = m.Seg.B.Add(offset)
}

func (m *SegTweenMover) Pos() v.Vec {
	return m.Seg.A.Lerp(m.Seg.B, m.Tw.Value)
}
