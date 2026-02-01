package ui

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
)

type DebugLayer struct {
	font    *Font
	visible bool

	fpsText      *Text
	positionText *Text
	chunkText    *Text
	facingText   *Text
	memText      *Text
	statsText    *Text
	targetText   *Text
}

func NewDebugLayer(font *Font, width, height int) *DebugLayer {
	return &DebugLayer{
		font:    font,
		visible: false,

		fpsText:      NewText(font, "FPS: 0", 10, 30, 0.5, mgl32.Vec3{1, 1, 0}), // Yellow
		positionText: NewText(font, "Pos: 0,0,0", 10, 50, 0.5, mgl32.Vec3{1, 1, 1}),
		chunkText:    NewText(font, "Chunk: 0,0", 10, 70, 0.5, mgl32.Vec3{1, 1, 1}),
		facingText:   NewText(font, "Facing: ?", 10, 90, 0.5, mgl32.Vec3{1, 1, 1}),
		memText:      NewText(font, "Mem: 0MB", 10, 110, 0.5, mgl32.Vec3{1, 1, 1}),
		statsText:    NewText(font, "Render: -", 10, 130, 0.5, mgl32.Vec3{1, 1, 1}),
		targetText:   NewText(font, "Target: -", 10, 150, 0.5, mgl32.Vec3{1, 1, 1}),
	}
}

func (d *DebugLayer) Init() error {
	// Initialize all text lines
	d.fpsText.Init()
	d.positionText.Init()
	d.chunkText.Init()
	d.facingText.Init()
	d.memText.Init()
	d.statsText.Init()
	d.targetText.Init()
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
	d.memText.Update(nil)
	d.statsText.Update(nil)
	d.targetText.Update(nil)
}

func (d *DebugLayer) Draw(shader uint32, proj mgl32.Mat4) {
	if !d.visible {
		return
	}

	d.fpsText.Draw(shader, proj)
	d.positionText.Draw(shader, proj)
	d.chunkText.Draw(shader, proj)
	d.facingText.Draw(shader, proj)
	d.memText.Draw(shader, proj)
	d.statsText.Draw(shader, proj)
	d.targetText.Draw(shader, proj)
}

func (d *DebugLayer) Cleanup() {
	d.fpsText.Cleanup()
	d.positionText.Cleanup()
	d.chunkText.Cleanup()
	d.facingText.Cleanup()
	d.memText.Cleanup()
	d.statsText.Cleanup()
	d.targetText.Cleanup()
}

func (d *DebugLayer) Toggle() bool {
	d.visible = !d.visible
	return d.visible
}

func (d *DebugLayer) UpdateInfo(fps float64,
	frameTime float32,
	pos mgl32.Vec3,
	facing mgl32.Vec3, // RENAMED from 'dir' to 'facing' to match your logic below
	chunkX, chunkZ int,
	memMB uint64,
	goroutines int,
	renderedChunks int,
	totalVerts int32,
	targetBlock string) {
	if !d.visible {
		return
	}
	d.fpsText.SetContent(fmt.Sprintf("FPS: %.0f (%.2f ms)", fps, frameTime*1000))
	d.positionText.SetContent(fmt.Sprintf("Pos: %.1f, %.1f, %.1f", pos.X(), pos.Y(), pos.Z()))
	d.chunkText.SetContent(fmt.Sprintf("Chunk: %d, %d", chunkX, chunkZ))

	directionStr := "North"
	if abs(facing.X()) > abs(facing.Z()) {
		if facing.X() > 0 {
			directionStr = "East"
		} else {
			directionStr = "West"
		}
	} else {
		if facing.Z() > 0 {
			directionStr = "South"
		} else {
			directionStr = "North"
		}
	}
	d.facingText.SetContent(fmt.Sprintf("Facing: %s", directionStr))

	d.memText.SetContent(fmt.Sprintf("Mem: %d MB | GRT: %d", memMB, goroutines))

	d.statsText.SetContent(fmt.Sprintf("Render: %d Chunks | %dk Verts", renderedChunks, totalVerts/1000))

	d.targetText.SetContent(fmt.Sprintf("Target: %s", targetBlock))
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
