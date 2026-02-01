package ui

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
)

type DebugLayer struct {
	font    *Font
	visible bool

	// The individual text lines
	fpsText      *Text
	positionText *Text
	chunkText    *Text
	facingText   *Text
}

func NewDebugLayer(font *Font, width, height int) *DebugLayer {
	return &DebugLayer{
		font:    font,
		visible: false,

		fpsText:      NewText(font, "FPS: 0", 10, 30, 0.5, mgl32.Vec3{1, 1, 0}),     // Green
		positionText: NewText(font, "Pos: 0,0,0", 10, 50, 0.5, mgl32.Vec3{1, 1, 1}), // White
		chunkText:    NewText(font, "Chunk: 0,0", 10, 70, 0.5, mgl32.Vec3{1, 1, 1}), // White
		facingText:   NewText(font, "Facing: ?", 10, 90, 0.5, mgl32.Vec3{1, 1, 1}),  // White
	}
}

func (d *DebugLayer) Init() error {
	// Initialize all text lines
	d.fpsText.Init()
	d.positionText.Init()
	d.chunkText.Init()
	d.facingText.Init()
	return nil
}

func (d *DebugLayer) Update(state interface{}) {
	if !d.visible {
		return
	}
	d.fpsText.Update(nil)
	d.positionText.Update(nil)
	d.chunkText.Update(nil)
	d.facingText.Update(nil)
}

func (d *DebugLayer) Draw(shader uint32, proj mgl32.Mat4) {
	if !d.visible {
		return
	}

	d.fpsText.Draw(shader, proj)
	d.positionText.Draw(shader, proj)
	d.chunkText.Draw(shader, proj)
	d.facingText.Draw(shader, proj)
}

func (d *DebugLayer) Cleanup() {
	d.fpsText.Cleanup()
	d.positionText.Cleanup()
	d.chunkText.Cleanup()
	d.facingText.Cleanup()
}

func (d *DebugLayer) Toggle() bool {
	d.visible = !d.visible
	return d.visible
}

func (d *DebugLayer) UpdateInfo(fps float64, pos mgl32.Vec3, chunkX, chunkZ int, facing mgl32.Vec3) {
	if !d.visible {
		return
	}

	d.fpsText.SetContent(fmt.Sprintf("FPS: %.0f", fps))
	d.positionText.SetContent(fmt.Sprintf("Pos: %.1f, %.1f, %.1f", pos.X(), pos.Y(), pos.Z()))
	d.chunkText.SetContent(fmt.Sprintf("Chunk: %d, %d", chunkX, chunkZ))

	dir := "North"
	if abs(facing.X()) > abs(facing.Z()) {
		if facing.X() > 0 {
			dir = "East"
		} else {
			dir = "West"
		}
	} else {
		if facing.Z() > 0 {
			dir = "South"
		} else {
			dir = "North"
		}
	}
	d.facingText.SetContent(fmt.Sprintf("Facing: %s", dir))
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
