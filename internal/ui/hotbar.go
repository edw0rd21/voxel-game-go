package ui

import (
	"voxel-game/internal/world"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Hotbar struct {
	vao uint32
	vbo uint32

	screenWidth  int
	screenHeight int

	selectedSlot int
	slotCount    int
	slotSize     float32
	padding      float32

	needsUpdate bool

	fillVertexCount   int
	borderVertexCount int

	texture uint32
}

func NewHotbar(screenWidth, screenHeight int) *Hotbar {
	return &Hotbar{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		selectedSlot: 0, // Dirt by default
		slotCount:    6, // 3 block types
		slotSize:     50.0,
		padding:      5.0,
		needsUpdate:  true,
	}
}

func (h *Hotbar) Init() error {
	gl.GenVertexArrays(1, &h.vao)
	gl.GenBuffers(1, &h.vbo)
	checkGLError("Hotbar.Init after creating VAO/VBO")

	// Create 1x1 White Texture
	gl.GenTextures(1, &h.texture)
	gl.BindTexture(gl.TEXTURE_2D, h.texture)
	white := []uint8{255, 255, 255, 255}
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(white))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	h.generateGeometry()
	return nil
}

func (h *Hotbar) generateGeometry() {
	// Calculate total width and starting X position
	totalWidth := float32(h.slotCount)*h.slotSize + float32(h.slotCount-1)*h.padding
	startX := (float32(h.screenWidth) - totalWidth) / 2.0
	bottomY := float32(h.screenHeight) - 80.0 // 80 pixels from bottom

	fillVertices := make([]float32, 0)
	borderVertices := make([]float32, 0)

	borderThickness := float32(2.0)

	// Draw slots
	for slotIndex := 0; slotIndex < h.slotCount; slotIndex++ {
		i := slotIndex
		x := startX + float32(i)*(h.slotSize+h.padding)

		// Determine color based on selection and block type
		var borderColor mgl32.Vec3
		if i == h.selectedSlot {
			borderColor = mgl32.Vec3{1.0, 1.0, 1.0} // White border for selected
		} else {
			borderColor = mgl32.Vec3{0.5, 0.5, 0.5} // Gray for unselected
		}

		// Get block color for the fill
		blockColor := getBlockColorForSlot(i)

		// Draw filled rectangle (block preview)
		innerPadding := float32(5.0)
		if i == h.selectedSlot {
			innerPadding = 3.0 // Less padding for selected
		}

		fillVertices = append(fillVertices, createFilledRect(
			x+innerPadding,
			bottomY+innerPadding,
			h.slotSize-innerPadding*2,
			h.slotSize-innerPadding*2,
			blockColor)...)

		// Draw border as 4 thin rectangles
		// Top border
		borderVertices = append(borderVertices, createFilledRect(
			x,
			bottomY,
			h.slotSize,
			borderThickness,
			borderColor)...)
		// Bottom border
		borderVertices = append(borderVertices, createFilledRect(
			x,
			bottomY+h.slotSize-borderThickness,
			h.slotSize,
			borderThickness,
			borderColor)...)
		// Left border
		borderVertices = append(borderVertices, createFilledRect(
			x, bottomY,
			borderThickness,
			h.slotSize,
			borderColor)...)
		// Right border
		borderVertices = append(borderVertices, createFilledRect(
			x+h.slotSize-borderThickness,
			bottomY,
			borderThickness,
			h.slotSize,
			borderColor)...)
	}

	h.fillVertexCount = len(fillVertices) / 7
	h.borderVertexCount = len(borderVertices) / 7
	stride := int32(7 * 4)

	// Upload to VBO
	gl.BindVertexArray(h.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, h.vbo)

	combined := append(fillVertices, borderVertices...)
	gl.BufferData(gl.ARRAY_BUFFER, len(combined)*4, gl.Ptr(combined), gl.DYNAMIC_DRAW)

	// Position attribute (2D)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, stride, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Color attribute
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	//Texture Coord
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(5*4))
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)

	checkGLError("Hotbar.generateGeometry")

	h.needsUpdate = false
}

func (h *Hotbar) Update(state interface{}) {
	if selectedBlock, ok := state.(world.BlockType); ok {
		newSlot := int(selectedBlock) - 1 // BlockDirt=1 -> slot 0
		if newSlot >= 0 && newSlot < h.slotCount && newSlot != h.selectedSlot {
			h.selectedSlot = newSlot
			h.needsUpdate = true
		}
	}

	if screenSize, ok := state.(*ScreenSize); ok {
		if screenSize.Width != h.screenWidth || screenSize.Height != h.screenHeight {
			h.screenWidth = screenSize.Width
			h.screenHeight = screenSize.Height
			h.needsUpdate = true
		}
	}

	// Regenerate geometry if needed
	if h.needsUpdate {
		h.generateGeometry()
	}
}

func (h *Hotbar) Draw(shaderProgram uint32, projection mgl32.Mat4) {
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, h.texture)

	gl.BindVertexArray(h.vao)
	// Draw fill batch
	gl.DrawArrays(gl.TRIANGLES, 0, int32(h.fillVertexCount))

	// Draw border batch
	gl.DrawArrays(gl.TRIANGLES, int32(h.fillVertexCount), int32(h.borderVertexCount))

	gl.BindVertexArray(0)

	checkGLError("Hotbar.Draw")
}

func (h *Hotbar) Cleanup() {
	gl.DeleteVertexArrays(1, &h.vao)
	gl.DeleteBuffers(1, &h.vbo)
}

func getBlockColorForSlot(slot int) mgl32.Vec3 {
	switch slot {
	case 0: // Dirt
		return mgl32.Vec3{0.6, 0.4, 0.2}
	case 1: // Grass
		return mgl32.Vec3{0.2, 0.8, 0.2}
	case 2: // Stone
		return mgl32.Vec3{0.5, 0.5, 0.5}
	case 3: // Snow
		return mgl32.Vec3{1.0, 1.0, 1.0}
	case 4: // Sand
		return mgl32.Vec3{0.9, 0.8, 0.6}
	case 5: // Wood
		return mgl32.Vec3{0.5, 0.3, 0.1}
	default:
		return mgl32.Vec3{1.0, 1.0, 1.0}
	}
}
