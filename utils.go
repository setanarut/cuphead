package main

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/setanarut/aseplayer"
	"github.com/setanarut/coll"
	"github.com/setanarut/v"
)

type OffsetToggler float64

func (t *OffsetToggler) Next() float64 {
	*t = -*t
	return float64(*t)
}

func CalculateGravityForHeight(height float64) float64 {
	return math.Pow(JumpPower, 2) * 0.5 / height
}

func CalculateMaxJumpHeight() float64 {
	return math.Pow(JumpPower, 2) * 0.5 / Gravity
}

var lastLine string
var count int

func SmartPrint(msg string) {
	if msg == lastLine {
		count++
		fmt.Printf("\r%s (%d)", msg, count)
	} else {
		if lastLine != "" {
			fmt.Println()
		}
		count = 1
		lastLine = msg
		fmt.Print(msg)
	}
}

func CursorScreen() v.Vec {
	x, y := ebiten.CursorPosition()
	return v.Vec{float64(x), float64(y)}
}

func slowMotionToggle(tps int) {
	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		switch ebiten.TPS() {
		case 60:
			ebiten.SetTPS(tps)
		case tps:
			ebiten.SetTPS(60)
		}
	}
}

func float64ToString(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}

func GetImageCenterOffset(i image.Image) (offset v.Vec) {
	offset.X = float64(i.Bounds().Dx())
	offset.Y = float64(i.Bounds().Dy())
	offset = offset.Scale(0.5).Neg()
	return
}

func resizeBoxFixedBottom(box *coll.AABB, newHalfY float64) {
	bottom := box.Bottom() // Mevcut alt kenar pozisyonunu kaydet
	box.Half.Y = newHalfY  // Yeni yarı yüksekliği ayarla
	box.SetBottom(bottom)  // Alt kenarı orijinal pozisyonuna geri ayarla
}
func resizeBoxFixedTop(box *coll.AABB, newHalfY float64) {
	top := box.Top()      // Mevcut alt kenar pozisyonunu kaydet
	box.Half.Y = newHalfY // Yeni yarı yüksekliği ayarla
	box.SetTop(top)       // Üst kenarı orijinal pozisyonuna geri ayarla
}

func rect(x, y, w, h int) image.Rectangle {
	return image.Rect(x, y, x+w, y+h)
}

func parseOffset(userData string) (pivot v.Vec) {
	parts := strings.Split(userData, ",")
	pivot.X, _ = strconv.ParseFloat(parts[0], 64)
	pivot.Y, _ = strconv.ParseFloat(parts[1], 64)
	return
}

func parseUserDataOffsets(ap *aseplayer.AnimPlayer) {
	for _, v := range ap.Animations {
		offset := parseOffset(v.UserData)
		for i := range v.Frames {
			v.Frames[i].Position.X -= offset.X
			v.Frames[i].Position.Y -= offset.Y
		}
	}
}

func readAseprite(filename string) *aseplayer.AnimPlayer {
	ase, _ := aseplayer.NewAnimPlayerFromAsepriteFileSystem(assets, "img/"+filename)
	parseUserDataOffsets(ase)
	return ase
}

func getFiles(dir string) []string {
	var fileNames []string
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == "" {
			fileNames = append(fileNames, file.Name())
		}
	}
	return fileNames
}

type Timer struct {
	Target  time.Duration
	Elapsed time.Duration
	tick    time.Duration
}

func NewTimer(duration time.Duration) Timer {

	return Timer{
		Target:  duration,
		Elapsed: 0,
		tick:    Tick,
	}

}

func (t *Timer) Update() {
	if t.Elapsed < t.Target {
		t.Elapsed += t.tick

	}
}

func (t *Timer) IsReady() bool {
	return t.Elapsed > t.Target
}
func (t *Timer) IsStart() bool {
	return t.Elapsed == 0
}

func (t *Timer) Reset() {
	t.Elapsed = 0
}

func (t *Timer) Remaining() time.Duration {
	return t.Target - t.Elapsed
}
func (t *Timer) RemainingSecondsString() string {
	return fmt.Sprintf("%.1fs", t.Remaining().Abs().Seconds())
}

// Pointer karşılaştırması ile sona taşı
func moveToEndByPtr[T any](slice []*T, ptr *T) []*T {
	index := slices.IndexFunc(slice, func(item *T) bool {
		return item == ptr // pointer adresi karşılaştırması
	})

	if index == -1 {
		return slice // bulunamadı, değişiklik yapma
	}

	slice = slices.Delete(slice, index, index+1)
	return append(slice, ptr)
}
