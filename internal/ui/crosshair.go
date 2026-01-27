package ui

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Crosshair struct {
	vao   uint32
	vbo   uint32
	color mgl32.Vec3
	size  float32
	
	screenWidth  int
	screenHeight int
}

func NewCrosshair(screenWidth, screenHeight int) *Crosshair {
	return &Crosshair{
		color:        mgl32.Vec3{1.0, 1.0, 1.0}, // White
		size:         10.0,                        // 10 pixels
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}
}

func (c *Crosshair) Init() error {
	// Generate VAO and VBO
	gl.GenVertexArrays(1, &c.vao)
	gl.GenBuffers(1, &c.vbo)

	// Generate geometry
	c.generateGeometry()

	return nil
}

func (c *Crosshair) generateGeometry() {
	// Calculate center of screen
	centerX := float32(c.screenWidth) / 2.0
	centerY := float32(c.screenHeight) / 2.0

	// Create crosshair lines (+ shape)
	vertices := []float32{
		// Horizontal line
		centerX - c.size, centerY, c.color[0], c.color[1], c.color[2],
		centerX + c.size, centerY, c.color[0], c.color[1], c.color[2],
		// Vertical line
		centerX, centerY - c.size, c.color[0], c.color[1], c.color[2],
		centerX, centerY + c.size, c.color[0], c.color[1], c.color[2],
	}

	gl.BindVertexArray(c.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, c.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Position attribute (2D)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Color attribute
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(2*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
}

func (c *Crosshair) Update(state interface{}) {
	// Crosshair doesn't change based on state
	// But we could update color, size, etc. here if needed
	
	// If screen size changed, regenerate
	if newSize, ok := state.(*ScreenSize); ok {
		if newSize.Width != c.screenWidth || newSize.Height != c.screenHeight {
			c.screenWidth = newSize.Width
			c.screenHeight = newSize.Height
			c.generateGeometry()
		}
	}
}

func (c *Crosshair) Draw(shaderProgram uint32, projection mgl32.Mat4) {
	gl.BindVertexArray(c.vao)
	gl.LineWidth(2.0)
	gl.DrawArrays(gl.LINES, 0, 4)
	gl.BindVertexArray(0)
}

func (c *Crosshair) Cleanup() {
	gl.DeleteVertexArrays(1, &c.vao)
	gl.DeleteBuffers(1, &c.vbo)
}

func (c *Crosshair) SetColor(color mgl32.Vec3) {
	c.color = color
	c.generateGeometry()
}

func (c *Crosshair) SetSize(size float32) {
	c.size = size
	c.generateGeometry()
}
