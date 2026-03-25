package main

import (
	"fmt"
	"image"
	"image/color"
	"slices"
	"strconv"
	"time"

	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/setanarut/coll"
	"github.com/setanarut/kamera/v2"
	"github.com/setanarut/tween"
	"github.com/setanarut/v"
	"golang.org/x/image/colornames"
)

type LevelEditor struct {
	editorSettingsWindowRect image.Rectangle
	objPropertyWindowRect    image.Rectangle
	selected                 any
	yGuide                   float64
	moverAngle               float64
	editorUi                 debugui.DebugUI
	editorSetting            debugui.DebugUI

	// Tıklama noktası ile platform merkezi arasındaki fark
	dragOffset v.Vec
	cursor     v.Vec

	playerHitBoxColor color.Color
	pathColor         color.Color
	pathImage         *ebiten.Image
	strokeW           float32
	strokeW64         float64

	bufString string

	easeList             []string
	cameraSmoothTypes    []string
	segmentTweenDuration int
	easeIndex            int
	cameraSmoothIndex    int

	showDashSweptBox bool
	showGuides       bool
	drawBodies       bool
	showPlayerHitbox bool

	isDragging        bool
	cursorJustPressed bool
	cursorPressed     bool
	cmdPressed        bool
	shiftPressed      bool
	resizePressed     bool
	gJustPressed      bool
	pauseJustPressed  bool

	focusObjectWindow  debugui.InputCapturingState
	focusEditortWindow debugui.InputCapturingState

	dio ebiten.DrawImageOptions
}

func (e *LevelEditor) Init(l ILevel) {
	e.cameraSmoothTypes = []string{"None", "Lerp", "SmoothDamp"}
	e.drawBodies = false
	e.easeList = []string{"InOutCubic", "Linear"}
	e.segmentTweenDuration = 1500
	e.editorSettingsWindowRect = rect(int(screenSize.X)-200, 0, 200, int(screenSize.Y)-20)
	e.objPropertyWindowRect = rect(0, 0, 200, int(screenSize.Y)-20)
	e.strokeW = 1
	e.strokeW64 = 1
	e.pathImage = ebiten.NewImage(1, 1)
	e.pathColor = color.RGBA{190, 122, 66, 255} // rgba(190, 122, 66, 1)
	e.pathImage.Fill(e.pathColor)
	e.dio = ebiten.DrawImageOptions{}

}

// Object Windoww
func (e *LevelEditor) objectWindow() error {
	if e.isSelected() {
		var err error
		if e.focusObjectWindow, err = e.editorUi.Update(func(ctx *debugui.Context) error {
			ctx.Window("Object", e.objPropertyWindowRect, func(layout debugui.ContainerLayout) {
				// ctx.Text("Position: " + ent.position().String())
				switch obj := e.selected.(type) {
				case *Player:
					ctx.Text(fmt.Sprintf("%T", obj.activeState))
					ctx.Text(fmt.Sprintf("%T", obj.grounded.activeSubState))
					ctx.Text(fmt.Sprintf("isOnFloor %v", obj.IsOnFloor))
					ctx.Text(fmt.Sprintf("isparry %v", obj.IsParrying))
					ctx.Text(fmt.Sprintf("hitTimer %v", obj.HitTimer))
				case *Body:
					ctx.TextField(&obj.Tag)
					ro := ctx.Button("Reset Offsets")
					ro.On(func() {
						obj.ResetSpriteOffsets()
					})

					ctx.Slider(&obj.ImgIndex, -1, 30, 1)
					ctx.Checkbox(&obj.StaticImage, "Static image")
					ctx.SliderF(&obj.ImgScale, 0.1, 2, 0.01, 2)
					ctx.SliderF(&obj.ImgOffset.X, -100, 200, 1, 2)
					ctx.SliderF(&obj.ImgOffset.Y, -100, 200, 1, 2)
					ctx.Checkbox(&obj.StaticBody, "Body Static")
					ctx.Text(fmt.Sprintf("Total Shapes %v", len(obj.Shapes)))
					ctx.Text("HP")
					ctx.NumberField(&obj.Hp, 1)
					ctx.Text("X")
					ctx.NumberFieldF(&obj.Pos.X, 1, 2)
					ctx.Text("Y")
					ctx.NumberFieldF(&obj.Pos.Y, 1, 2)

					switch mover := obj.Mover.(type) {
					case *OrbitalMover:
						ctx.Header("Orbital Mover", true, func() {
							ctx.Text("Orbit Radius")
							ctx.SliderF(&mover.Orbit.Radius, 10, 100, 1, 0)
							ctx.Text("Orbit Speed")
							ctx.SliderF(&mover.Speed, 0, 0.2, 0.0001, 5)
						})
					case *SegTweenMover:
						ctx.Checkbox(&mover.Tw.Yoyo, "yoyo")
						ctx.Checkbox(&mover.Tw.Reversed, "reversed")
						ctx.Header("Seg tween mover", true, func() {
							ease := ctx.Dropdown(&e.easeIndex, e.easeList)
							ease.On(func() {
								mover.Tw.EasingFunc = tween.EaseMap[e.easeList[e.easeIndex]]
							})

							ctx.Text("A XY")
							ctx.NumberFieldF(&mover.Seg.A.X, 1, 2)
							ctx.NumberFieldF(&mover.Seg.A.Y, 1, 2)
							ctx.Text("B XY")
							ctx.NumberFieldF(&mover.Seg.B.X, 1, 2)
							ctx.NumberFieldF(&mover.Seg.B.Y, 1, 2)

							sl := ctx.Slider(&e.segmentTweenDuration, 500, 10000, 1)
							sl.On(func() {
								mover.Tw.Duration = time.Duration(e.segmentTweenDuration) * time.Millisecond
							})

						})
					case *PathMover:
						ctx.Header("Path Mover", true, func() {
							ctx.Text(strconv.Itoa(mover.Path.Len()) + " Points")
							eh := ctx.NumberFieldF(&e.moverAngle, 0.01, 2)
							eh.On(func() {
								mover.Path.Rotate(e.moverAngle)
							})
						})
					}

					onb := ctx.Button("Delete Mover")
					onb.On(func() {
						obj.Mover = nil
					})

				case *BoxShape:
					ctx.Text("BoxShape")
					ctx.Text(fmt.Sprintf("Parent tag: %s", obj.parent.Tag))
					ctx.Checkbox(&obj.IgnoreBullet, "ignore bullet")
					ctx.Checkbox(&obj.OneWay, "OneWay")
					ctx.Checkbox(&obj.HorizontalCollisionDisabled, "HorizontalCollisionDisabled")
					ctx.Checkbox(&obj.Solid, "Solid")
					ctx.Checkbox(&obj.Sensor, "Sensor")
					ctx.Checkbox(&obj.Parry, "Parry")
					ctx.Checkbox(&obj.CanGiveDamage, "CanGiveDamage")
					ctx.Checkbox(&obj.CanTakeBulletDamage, "CanTakeBulletDamage")
					ctx.Checkbox(&obj.CanTakePlayerDamage, "CanTakePlayerDamage")

					ctx.Text(fmt.Sprintf("Pos %v", obj.Pos))
					ctx.Text("half")
					ctx.NumberFieldF(&obj.Half.X, 1, 2)
					ctx.NumberFieldF(&obj.Half.Y, 1, 2)
					ctx.Text("local offset")
					ctx.NumberFieldF(&obj.LocalOffset.X, 1, 2)
					ctx.NumberFieldF(&obj.LocalOffset.Y, 1, 2)

				case *CircleShape:
					ctx.Text("CircleShape")
					ctx.Text(fmt.Sprintf("Parent tag: %s", obj.parent.Tag))
					ctx.SliderF(&obj.Radius, 3, 100, 1, 0)
					ctx.Checkbox(&obj.IgnoreBullet, "ignore bullet")
					ctx.Checkbox(&obj.OneWay, "OneWay")
					ctx.Checkbox(&obj.HorizontalCollisionDisabled, "HorizontalCollisionDisabled")
					ctx.Checkbox(&obj.Solid, "Solid")
					ctx.Checkbox(&obj.Parry, "Parry")
					ctx.Checkbox(&obj.CanGiveDamage, "CanGiveDamage")
					ctx.Checkbox(&obj.CanTakeBulletDamage, "CanTakeBulletDamage")
					ctx.Checkbox(&obj.CanTakePlayerDamage, "CanTakePlayerDamage")
					ctx.Text("local offset")
					ctx.NumberFieldF(&obj.LocalOffset.X, 1, 2)
					ctx.NumberFieldF(&obj.LocalOffset.Y, 1, 2)

				}
			},
			)
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

// Editor Window
func (e *LevelEditor) editorWindow(lvl ILevel) error {
	var err error
	if e.focusEditortWindow, err = e.editorSetting.Update(func(ctx *debugui.Context) error {
		ctx.Window("Editor Settings", e.editorSettingsWindowRect, func(layout debugui.ContainerLayout) {
			ctx.Text(fmt.Sprintf("selected Type %T", e.selected))
			front := ctx.Button("Move to Front")
			front.On(func() {
				bd := *lvl.Bodies()
				moveToEndByPtr(bd, e.selected.(*Body))
			})
			ctx.Text("Cuphead HP")
			ctx.NumberField(&Cuphead.Hp, 1)

			ctx.Checkbox(&e.drawBodies, "draw bodies")
			ctx.Checkbox(&e.showPlayerHitbox, "Show player hitBox")
			ctx.Checkbox(&GamePaused, "Game Paused")
			ctx.Checkbox(&Cuphead.Paused, "Cuphead Paused")
			ctx.Text("Fill Color")
			ctx.Slider(&screenFillColor, 0, 255, 1)
			ctx.SliderF(&e.strokeW64, 1, 10, 0.5, 2)

			ctx.Header("Camera", false, func() {
				// ctx.Text(MainCamera.String())
				// ctx.Checkbox(&e.level.StaticCamera, "Static Camera")
				eve := ctx.SliderF(&CameraTarget.Y, 0, 500, 1, 2)
				eve.On(func() {
					MainCamera.SetCenter(Cuphead.Pos.X, CameraTarget.Y)
				})
				ev := ctx.Dropdown(&e.cameraSmoothIndex, e.cameraSmoothTypes)
				ev.On(func() {
					MainCamera.SmoothType = kamera.SmoothType(e.cameraSmoothIndex)
				})
			})

			ctx.Header("Save/Load Level", true, func() {
				saveBut := ctx.Button("Save level")
				saveBut.On(func() {
					lvl.Save()
				})
				load := ctx.Button("Load level")
				load.On(func() {
					lvl.Load()
				})
			})
		},
		)
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (e *LevelEditor) Update(lvl ILevel) error {

	e.strokeW = float32(e.strokeW64)

	e.inputs()

	if e.pauseJustPressed && e.cmdPressed {
		GamePaused = !GamePaused
	}

	// Pause
	if inpututil.IsKeyJustPressed(ebiten.KeyP) && e.shiftPressed {
		GameFreeze = !GameFreeze
	}

	// teleport cuphead
	if inpututil.IsKeyJustPressed(ebiten.KeyIntlBackslash) {
		Cuphead.Teleport(e.cursor)
	}

	// rehber çizgisi taşı
	if e.gJustPressed {
		e.yGuide = e.cursor.Y
	}

	if !e.cursorPressed {
		e.isDragging = false
	}

	e.editorWindow(lvl)
	e.objectWindow()

	if e.focusEditortWindow == 0 && e.focusObjectWindow == 0 {
		e.selectObject(lvl)
		e.deleteObject(lvl)
		e.addObject(lvl)
		e.duplicateObject(lvl)
		e.addSCirclehape()
		e.resize()
		e.dragObject()
		slowMotionToggle(6)
	}
	return nil
}

func (e *LevelEditor) resize() {
	if e.cursorPressed && e.resizePressed {

		switch p := e.selected.(type) {
		case *Body:
			for _, s := range p.Shapes {
				switch shape := s.(type) {
				case *BoxShape:
					shape.Half = e.cursor.Sub(shape.Pos).Abs()
				case *CircleShape:
					shape.Radius = e.cursor.Dist(shape.Pos)
				}
			}
		}
	}
}

func (e *LevelEditor) selectObjectWithTag(tag string, lvl ILevel) {

	i := slices.IndexFunc(*lvl.Bodies(), func(b *Body) bool {
		return b.Tag == tag
	})

	if i != -1 {
		e.selected = (*lvl.Bodies())[i]
	}

}
func (e *LevelEditor) selectObject(lvl ILevel) {
	if e.cursorJustPressed {
		for _, b := range *lvl.Bodies() {
			for _, s := range b.Shapes {
				switch shape := s.(type) {
				case *BoxShape:
					// if b != e.selected {
					if coll.BoxPointOverlap(&shape.AABB, e.cursor, nil) {
						if e.shiftPressed {
							e.selected = s
						} else {
							e.selected = b
						}
						break
					}
					// }
				case *CircleShape:
					// if b != e.selected {
					if coll.CirclePointOverlap(&shape.Circle, e.cursor, nil) {
						if e.shiftPressed {
							e.selected = s
						} else {
							e.selected = b
						}
						break
					}
					// }
				}
			}
		}

		if coll.BoxPointOverlap(&Cuphead.AABB, e.cursor, nil) {
			e.selected = Cuphead
		}

	}
}

func (e *LevelEditor) dragObject() {

	// offset hesapla - sadece imleç platformun içindeyse
	if e.cursorJustPressed && !e.resizePressed {
		if e.isSelected() {
			switch obj := e.selected.(type) {
			case *Body:
				for _, s := range obj.Shapes {
					switch shape := s.(type) {
					case *BoxShape:
						if coll.BoxPointOverlap(&shape.AABB, e.cursor, nil) {
							e.dragOffset = obj.Pos.Sub(e.cursor)
							e.isDragging = true
						}
					case *CircleShape:
						if coll.CirclePointOverlap(&shape.Circle, e.cursor, nil) {
							e.dragOffset = obj.Pos.Sub(e.cursor)
							e.isDragging = true
						}
					}
				}
			}
		}
	}

	// taşı - sadece isDragging true ise
	if e.cursorPressed && !e.resizePressed && e.isDragging {
		if e.isSelected() && !e.cmdPressed {
			switch obj := e.selected.(type) {
			case *Body:
				obj.Pos = e.cursor.Add(e.dragOffset).Floor()

				if obj.Mover != nil {
					obj.Mover.SetPos(e.cursor.Add(e.dragOffset).Floor())
				}

				// for _, s := range obj.Shapes {
				// 	switch shape := s.(type) {
				// 	case *BoxShape:
				// 		box := &shape.AABB
				// 		box.Pos = e.cursor.Add(e.dragOffset).Floor()
				// 	case *CircleShape:
				// 		circ := &shape.Circle
				// 		circ.Pos = e.cursor.Add(e.dragOffset).Floor()
				// 	}
				// }
			}
		}
	}

	// taşı - sadece isDragging true ise
	if e.cursorJustPressed && !e.resizePressed && e.cmdPressed {
		if e.isSelected() {
			switch s := e.selected.(type) {
			case *CircleShape:
				s.LocalOffset = e.cursor.Sub(s.parent.Pos)
			}
		}
	}
}

func (e *LevelEditor) addObject(lvl ILevel) {

	// add box Body static
	if inpututil.IsKeyJustPressed(ebiten.Key1) && e.cmdPressed {
		b := NewBody(e.cursor, 3, true)
		b.AddShape(NewBoxShape(b, v.Vec{64, 16}, true))
		// b.AddShape(NewCircleShape(b, 12, false))
		bd := lvl.Bodies()
		*bd = append(*bd, b)
	}
	// add box Body dinamik
	if inpututil.IsKeyJustPressed(ebiten.Key2) && e.cmdPressed {
		b := NewBody(e.cursor, 3, false)
		b.AddShape(NewBoxShape(b, v.Vec{64, 16}, true))
		bd := lvl.Bodies()
		*bd = append(*bd, b)
	}
	// add circle Body dinamik seg
	if inpututil.IsKeyJustPressed(ebiten.Key3) && e.cmdPressed {
		b := NewBody(e.cursor, 3, false)
		s := NewCircleShape(b, 10, true, v.Vec{})
		b.AddShape(s)
		bd := lvl.Bodies()
		*bd = append(*bd, b)
	}
	// add Parry circle Body statik
	if inpututil.IsKeyJustPressed(ebiten.Key4) && e.cmdPressed {
		b := NewBody(e.cursor, 3, true)
		s := NewCircleShape(b, 20, true, v.Vec{0, 0})
		s.Parry = true
		b.AddShape(s)
		bd := lvl.Bodies()
		*bd = append(*bd, b)
	}

}
func (e *LevelEditor) duplicateObject(lvl ILevel) {
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) && e.cmdPressed {
		if e.isSelected() {
			if body, ok := e.selected.(*Body); ok {
				clone := body.Clone()
				clone.Pos.X += 10
				clone.Pos.Y += 10

				bd := lvl.Bodies()
				*bd = append(*bd, clone)

				e.selected = clone
			}
		}
	}
}
func (e *LevelEditor) addSCirclehape() {
	if inpututil.IsKeyJustPressed(ebiten.KeyV) && e.cmdPressed {
		if e.isSelected() {
			if body, ok := e.selected.(*Body); ok {
				body.AddShape(NewCircleShape(body, 10, true, v.Vec{}))
			}
		}
	}
}

func (e *LevelEditor) deleteObject(lvl ILevel) {
	if e.cmdPressed && inpututil.IsKeyJustPressed(ebiten.KeyX) {
		if e.isSelected() {
			switch selected := e.selected.(type) {
			case *CircleShape:
				bd := selected.parent
				bd.Shapes = slices.DeleteFunc(bd.Shapes, func(c IShape) bool {
					return c == selected
				})
			case *BoxShape:
				bd := selected.parent
				bd.Shapes = slices.DeleteFunc(bd.Shapes, func(c IShape) bool {
					return c == selected
				})
			case *Body:
				bd := lvl.Bodies()
				*bd = slices.DeleteFunc(*bd, func(p *Body) bool {
					return p == selected
				})
			}
		}
		e.selected = nil
	}
}

func (e *LevelEditor) inputs() {
	e.cursorJustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	e.cursorPressed = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	e.cmdPressed = ebiten.IsKeyPressed(ebiten.KeyMeta)
	e.shiftPressed = ebiten.IsKeyPressed(ebiten.KeyShift)
	e.resizePressed = ebiten.IsKeyPressed(ebiten.KeyC)
	e.gJustPressed = inpututil.IsKeyJustPressed(ebiten.KeyG)
	e.pauseJustPressed = inpututil.IsKeyJustPressed(ebiten.KeyP)
	e.cursor = e.cursorWorldPosition()
}

func (e *LevelEditor) Draw(screen *ebiten.Image, lvl ILevel) {

	if e.drawBodies {
		// bodies
		for _, body := range *lvl.Bodies() {
			e.drawBody(screen, body)
		}

		// Bullets
		for _, bullet := range ShootManag.Bullets {
			e.strokeCircle(screen, bullet.Circle, colornames.Yellow, e.strokeW)
		}

	}

	// Selected highlight
	e.drawHighLight(screen)

	// Player
	if e.showPlayerHitbox {
		switch Cuphead.activeState {
		case Cuphead.dash, Cuphead.jumping:
			e.playerHitBoxColor = colornames.Lightgreen
		default:
			e.playerHitBoxColor = colornames.Red
		}
		if Cuphead == e.selected {
			e.playerHitBoxColor = colornames.Yellow
		}

		if e.showDashSweptBox {
			checkBox := Cuphead.AABB
			checkBox.Pos.X += Cuphead.Facing * DashWallBlockDistance
			e.strokeAABB(&checkBox, colornames.Purple, screen, e.strokeW)
		}

		// Player Hit box
		e.strokeAABB(&Cuphead.AABB, e.playerHitBoxColor, screen, e.strokeW)

		// Parry sensor
		if Cuphead.activeState == Cuphead.fall {
			if Cuphead.IsParrying {
				e.strokeCircle(screen, Cuphead.ParrySensor, colornames.Orchid, e.strokeW)
			} else {
				e.strokeCircle(screen, Cuphead.ParrySensor, colornames.Darkslategray, e.strokeW)
			}
		}
	}

	if e.showGuides {
		e.strokePlane(screen, e.yGuide, colornames.Red)
	}
	e.editorUi.Draw(screen)
	e.editorSetting.Draw(screen)
}

func (e *LevelEditor) drawBody(screen *ebiten.Image, body *Body) {

	if body == nil {
		return
	}

	for _, s := range body.Shapes {

		clr := color.RGBA{114, 114, 114, 128} // rgb(193, 193, 193)

		if body.DamageFlashTimer > 0 {
			clr = color.RGBA{193, 193, 193, 255}
		}

		sd := s.GetShapeData()

		if sd.Parry {
			clr = color.RGBA{166, 64, 166, 255} // rgb(166, 64, 166)
		}

		switch shape := s.(type) {
		case *BoxShape:
			if shape.Solid {
				e.fillAABB(&shape.AABB, clr, screen)
			} else {
				e.strokeAABB(&shape.AABB, clr, screen, e.strokeW)
			}
		case *CircleShape:
			if shape.Solid || shape.Parry {
				e.fillCircleAt(screen, shape.Circle.Pos, shape.Circle.Radius, clr)
			} else {
				e.strokeCircle(screen, shape.Circle, clr, e.strokeW)
			}

		}
	}
}

func (e *LevelEditor) drawHighLight(screen *ebiten.Image) {
	if e.isSelected() {
		switch body := e.selected.(type) {
		case *Body:
			e.fillCircleAt(screen, body.Pos, 2, colornames.Cyan)
			for _, s := range body.Shapes {
				clr := color.RGBA{0, 132, 255, 255} // rgb(0, 132, 255)
				switch shape := s.(type) {
				case *BoxShape:
					e.strokeAABB(&shape.AABB, clr, screen, e.strokeW)
				case *CircleShape:
					e.strokeCircle(screen, shape.Circle, clr, e.strokeW)
				}
			}
			if body.Mover != nil {
				e.strokeCircleAt(screen, body.Mover.Pos(), 3, e.pathColor, e.strokeW)
				switch mover := body.Mover.(type) {
				case *SegTweenMover:
					e.strokeSegment(screen, &mover.Seg, e.pathColor)
				case *OrbitalMover:
					e.strokeCircle(screen, *mover.Orbit, e.pathColor, e.strokeW)
				case *PathMover:
					for _, point := range mover.Path.Points {
						e.dio.GeoM.Reset()
						e.dio.GeoM.Translate(point.X, point.Y)
						MainCamera.Draw(e.pathImage, &e.dio, screen)
					}
				}
			}
		}
	}
}

func (e *LevelEditor) isSelected() bool {
	return e.selected != nil
}

func (e *LevelEditor) strokePlatform(box *coll.AABB, dst *ebiten.Image, ow, hod bool, wd float32) {
	x, y := MainCamera.ApplyCameraTransformToPoint(box.Left(), box.Top())
	w := box.Width()
	h := box.Height()
	vector.StrokeLine(dst, float32(x), float32(y), float32(x+w), float32(y), wd, colornames.Dodgerblue, false)
	if ow {
		vector.StrokeLine(dst, float32(x), float32(y+h), float32(x+w), float32(y+h), wd, colornames.Forestgreen, false)
	} else {
		vector.StrokeLine(dst, float32(x), float32(y+h), float32(x+w), float32(y+h), wd, colornames.Darkgray, false)
	}
	if hod {
		vector.StrokeLine(dst, float32(x), float32(y), float32(x), float32(y+h), wd, colornames.Forestgreen, false)
		vector.StrokeLine(dst, float32(x+w), float32(y), float32(x+w), float32(y+h), wd, colornames.Forestgreen, false)
	} else {
		vector.StrokeLine(dst, float32(x), float32(y), float32(x), float32(y+h), wd, colornames.Darkgray, false)
		vector.StrokeLine(dst, float32(x+w), float32(y), float32(x+w), float32(y+h), wd, colornames.Darkgray, false)
	}
}

func (e *LevelEditor) strokeAABB(box *coll.AABB, clr color.Color, dst *ebiten.Image, strokeW float32) {
	x, y := MainCamera.ApplyCameraTransformToPoint(box.Left(), box.Top())
	vector.StrokeRect(
		dst,
		float32(x),
		float32(y),
		float32(box.Width()),
		float32(box.Height()),
		strokeW,
		clr,
		false,
	)
}
func (e *LevelEditor) fillAABB(box *coll.AABB, clr color.Color, dst *ebiten.Image) {
	x, y := MainCamera.ApplyCameraTransformToPoint(box.Left(), box.Top())
	vector.FillRect(
		dst,
		float32(x),
		float32(y),
		float32(box.Width()),
		float32(box.Height()),
		clr,
		false,
	)
}

func (e *LevelEditor) strokeSegment(s *ebiten.Image, seg *coll.Segment, clr color.Color) {
	ax, ay := MainCamera.ApplyCameraTransformToPoint(seg.A.X, seg.A.Y)
	bx, by := MainCamera.ApplyCameraTransformToPoint(seg.B.X, seg.B.Y)
	vector.StrokeLine(s, float32(ax), float32(ay), float32(bx), float32(by), 1, clr, false)
}
func (e *LevelEditor) strokePlane(s *ebiten.Image, y float64, clr color.Color) {
	_, ny := MainCamera.ApplyCameraTransformToPoint(0, y)
	vector.StrokeLine(s, float32(-3000), float32(ny), float32(3000), float32(ny), 1, clr, false)
}

func (e *LevelEditor) strokeCircle(dst *ebiten.Image, c coll.Circle, clr color.Color, sw float32) {
	x, y := MainCamera.ApplyCameraTransformToPoint(c.Pos.X, c.Pos.Y)
	vector.StrokeCircle(
		dst,
		float32(x),
		float32(y),
		float32(c.Radius),
		sw,
		clr,
		true,
	)
}

func (e *LevelEditor) fillCircleAt(dst *ebiten.Image, c v.Vec, radius float64, clr color.Color) {

	x, y := MainCamera.ApplyCameraTransformToPoint(c.X, c.Y)
	vector.FillCircle(dst, float32(x), float32(y), float32(radius), clr, true)
}

func (e *LevelEditor) strokeCircleAt(dst *ebiten.Image, c v.Vec, radius float64, clr color.Color, sw float32) {
	x, y := MainCamera.ApplyCameraTransformToPoint(c.X, c.Y)
	vector.StrokeCircle(dst, float32(x), float32(y), float32(radius), sw, clr, true)
}

func (e *LevelEditor) cursorWorldPosition() v.Vec {
	x, y := ebiten.CursorPosition()
	wx, wy := MainCamera.ScreenToWorld(x, y)
	return v.Vec{float64(wx), float64(wy)}
}
