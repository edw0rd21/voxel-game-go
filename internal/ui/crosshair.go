package ui

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Crosshair struct {
	vao       uint32
	vbo       uint32
	color     mgl32.Vec3
	size      float32
	thickness float32

	screenWidth  int
	screenHeight int

	vertexCount int
}

func NewCrosshair(screenWidth, screenHeight int) *Crosshair {
	return &Crosshair{
		color:        mgl32.Vec3{1.0, 1.0, 1.0}, // White
		size:         10.0,
		thickness:    2.0,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}
}

func (c *Crosshair) Init() error {
	// Generate VAO and VBO
	gl.GenVertexArrays(1, &c.vao)
	gl.GenBuffers(1, &c.vbo)

	checkGLError("Crosshair.Init after creating VAO/VBO")

	c.generateGeometry()

	return nil
}

func (c *Crosshair) generateGeometry() {
	// Calculate center of screen
	centerX := float32(c.screenWidth) / 2.0
	centerY := float32(c.screenHeight) / 2.0

	vertices := make([]float32, 0)

	//Horizontal and vertical lines
	vertices = append(vertices, createFilledRect(
		centerX-c.size,
		centerY-c.thickness/2,
		c.size*2,
		c.thickness,
		c.color,
	)...)

	vertices = append(vertices, createFilledRect(
		centerX-c.thickness/2,
		centerY-c.size,
		c.thickness,
		c.size*2,
		c.color,
	)...)

	c.vertexCount = len(vertices) / 7
	stride := int32(7 * 4)

	gl.BindVertexArray(c.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, c.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

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

	checkGLError("Crosshair.generateGeometry")
}

func (c *Crosshair) Update(state interface{}) {
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
	gl.DrawArrays(gl.TRIANGLES, 0, int32(c.vertexCount))
	gl.BindVertexArray(0)

	checkGLError("Crosshair.Draw")
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

func (c *Crosshair) SetThickness(thickness float32) {
	c.thickness = thickness
	c.generateGeometry()
}
