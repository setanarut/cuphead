package main

import (
	"encoding/gob"
	"fmt"
	"math"
	"os"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/setanarut/aseplayer"
	"github.com/setanarut/kamera/v2"
	"github.com/setanarut/mathutils"
	"github.com/setanarut/tween"
	"github.com/setanarut/v"
)

type TutorialLevel struct {
	CupheadStarPos v.Vec
	StaticCamera bool
	BodyList     *[]*Body

	spawnX     float64
	bodyMap    map[string]*Body
	tuto       *aseplayer.AnimPlayer
	statics    aseplayer.AnimPlayer
	staticAnim *aseplayer.Animation
	dio        *colorm.DrawImageOptions
	clrm       colorm.ColorM
}

func (t *TutorialLevel) Reset() {
	CameraTarget = v.Vec{}
	Cuphead.Teleport(t.CupheadStarPos)
	CameraTarget.X = Cuphead.Pos.X
	MainCamera.SetCenter(t.CupheadStarPos.X, CameraTarget.Y)
	MainCamera.SmoothType = kamera.Lerp
	t.StaticCamera = false
}

func (t *TutorialLevel) Init() {

	t.BodyList = &[]*Body{}
	t.bodyMap = make(map[string]*Body)

	t.dio = &colorm.DrawImageOptions{}
	t.clrm = colorm.ColorM{}
	t.dio.Filter = ebiten.FilterLinear

	t.tuto = readAseprite("tuto.ase")
	t.tuto.Play("ghost")

	t.statics = *t.tuto
	t.staticAnim = t.statics.Animations["static"]

	t.statics.Paused = true

	t.Reset()
	t.Load()

}

func (t *TutorialLevel) Update() {

	for _, body := range *t.BodyList {
		body.Anim.Update(Tick)
	}

	if Cuphead.Pos.X > t.spawnX {
		mover := t.bodyMap["ghost"].Mover.(*SegTweenMover)
		if mover.Tw.IsFinished() {
			mover.Tw.Reset()
		}
	}

	MainCamera.LookAt(Cuphead.Pos.X, CameraTarget.Y)
	MainCamera.X = math.Floor(MainCamera.X) + mathutils.Fract(Cuphead.Pos.X)

}

func (t *TutorialLevel) Draw(s *ebiten.Image) {
	for _, body := range *t.BodyList {
		t.dio.GeoM.Reset()
		t.clrm.Reset()
		if body.DamageFlashTimer > 0 {
			t.clrm.Translate(0.1, 0.1, 0.1, 0)
		}
		if body.StaticImage && body.ImgIndex != -1 && body.ImgIndex < len(t.staticAnim.Frames) {
			t.drawBodyWithStaticFrame(body, t.staticAnim, s)
		} else if !body.StaticImage {
			t.drawBodyWithAnim(body, s)
		}
	}
}

func (t *TutorialLevel) drawBodyWithStaticFrame(body *Body, static *aseplayer.Animation, s *ebiten.Image) {
	frame := static.Frames[body.ImgIndex]
	t.dio.GeoM.Translate(frame.Position.X, frame.Position.Y)
	t.dio.GeoM.Scale(body.ImgScale, body.ImgScale)
	t.dio.GeoM.Translate(body.Pos.X, body.Pos.Y)
	t.dio.GeoM.Translate(body.ImgOffset.X, body.ImgOffset.Y)
	MainCamera.DrawWithColorM(frame.Image, t.clrm, t.dio, s)
}

func (t *TutorialLevel) drawBodyWithAnim(body *Body, s *ebiten.Image) {
	t.dio.GeoM.Translate(body.Anim.CurrentFrame.Position.X, body.Anim.CurrentFrame.Position.Y)
	t.dio.GeoM.Scale(body.ImgScale, body.ImgScale)
	t.dio.GeoM.Translate(body.Pos.X, body.Pos.Y)
	t.dio.GeoM.Translate(body.ImgOffset.X, body.ImgOffset.Y)
	MainCamera.DrawWithColorM(body.Anim.CurrentFrame.Image, t.clrm, t.dio, s)
}

func (t *TutorialLevel) OnDamage(b *Body) {
	fmt.Println(b.Tag + " damaged")
	if b.Tag == "target" {
		t.bodyMap["kule"].DamageFlash()
	}
}

func (t *TutorialLevel) OnParry(b *Body) {

	shapeData := b.Shapes[0].GetShapeData()

	switch b.Tag {
	case "ghost":
		t.bodyMap["ghost"].Mover.(*SegTweenMover).Tw.SetTime(0)
	case "parry1":
		b.Anim.Play("sphere_gray")
		shapeData.Parry = false
		shapeData.Solid = false
		t.bodyMap["parry2"].Anim.Play("sphere")
		t.bodyMap["parry2"].Shapes[0].GetShapeData().Parry = true
		t.bodyMap["parry2"].Shapes[0].GetShapeData().Solid = true
	case "parry2":
		b.Anim.Play("sphere_gray")
		shapeData.Parry = false
		shapeData.Solid = false
		t.bodyMap["parry3"].Anim.Play("sphere")
		t.bodyMap["parry3"].Shapes[0].GetShapeData().Parry = true
		t.bodyMap["parry3"].Shapes[0].GetShapeData().Solid = true
	case "parry3":
		b.Anim.Play("sphere_gray")
		shapeData.Parry = false
		shapeData.Solid = false
		t.bodyMap["parry1"].Anim.Play("sphere")
		t.bodyMap["parry1"].Shapes[0].GetShapeData().Parry = true
		t.bodyMap["parry1"].Shapes[0].GetShapeData().Solid = true
	}
}

func (t *TutorialLevel) OnBodyRemove(b *Body) {
	if b.Tag == "target" {
		fmt.Println(b.Tag, "Removed")
		for _, v := range *t.BodyList {
			if v.Tag == "kule" {
				v.Hp = 0
			}
		}
	}

}

func (t *TutorialLevel) OnSensorEntered(b *Body) {
}

func (t *TutorialLevel) Bodies() *[]*Body {
	return t.BodyList
}

func (t *TutorialLevel) Save() {

	t.CupheadStarPos = Cuphead.Pos

	file, err := os.Create("levels/tuto")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	if err := enc.Encode(t); err != nil {
		panic(err)
	}

}

func (t *TutorialLevel) Load() {
	file, err := os.Open("levels/tuto")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	dec := gob.NewDecoder(file)

	if err := dec.Decode(t); err != nil {
		panic(err)
	}
	for _, b := range *t.BodyList {
		for _, s := range b.Shapes {
			if sd := s.GetShapeData(); sd != nil {
				sd.parent = b
			}
		}
		if b.Mover != nil {
			if sm, ok := b.Mover.(*SegTweenMover); ok {
				sm.Tw.EasingFunc = tween.EaseMap[sm.Tw.EaseName]
			}
		}
	}

	t.Reset()
	t.FillBodyMap()

	t.bodyMap["target"].StaticImage = false
	t.spawnX = t.bodyMap["parry_sensor"].Shapes[0].(*BoxShape).Left()
	t.bodyMap["parry1"].Anim = *t.tuto
	t.bodyMap["parry2"].Anim = *t.tuto
	t.bodyMap["parry3"].Anim = *t.tuto
	t.bodyMap["ghost"].Anim = *t.tuto
	t.bodyMap["target"].Anim = *t.tuto

	t.bodyMap["parry1"].Anim.Play("sphere")
	t.bodyMap["parry2"].Anim.Play("sphere_gray")
	t.bodyMap["parry3"].Anim.Play("sphere_gray")
	t.bodyMap["ghost"].Anim.Play("ghost")
	t.bodyMap["target"].Anim.Play("target")

	t.bodyMap["ghost"].StaticBody = false
	t.bodyMap["ghost"].StaticImage = false

	mv := t.bodyMap["ghost"].Mover.(*SegTweenMover)
	mv.Tw.Yoyo = false
	mv.Tw.Reversed = false
	mv.Tw.EasingFunc = tween.EaseMap[tween.Linear]

}

func (t *TutorialLevel) FillBodyMap() {
	clear(t.bodyMap)

	validTags := []string{"parry1", "parry2", "parry3", "parry_sensor", "ghost", "target", "kule"}
	for _, body := range *t.BodyList {
		body.Anim = t.statics
		if slices.Contains(validTags, body.Tag) {
			t.bodyMap[body.Tag] = body
		}
	}

}
