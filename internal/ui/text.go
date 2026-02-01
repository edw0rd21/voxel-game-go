package ui

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Text struct {
	font    *Font
	content string

	x, y  float32
	scale float32
	color mgl32.Vec3

	vao         uint32
	vbo         uint32
	vertexCount int32
	needsUpdate bool
}

func NewText(font *Font, content string, x, y float32, scale float32, color mgl32.Vec3) *Text {
	return &Text{
		font:        font,
		content:     content,
		x:           x,
		y:           y,
		scale:       scale,
		color:       color,
		needsUpdate: true,
	}
}

func (t *Text) Init() error {
	gl.GenVertexArrays(1, &t.vao)
	gl.GenBuffers(1, &t.vbo)

	gl.BindVertexArray(t.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, t.vbo)

	stride := int32(7 * 4) // 7 floats (X,Y, R,G,B, U,V) * 4 bytes

	// Position
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, stride, gl.PtrOffset(0))

	// Color
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, stride, gl.PtrOffset(2*4))

	// Texture Coords
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 2, gl.FLOAT, false, stride, gl.PtrOffset(5*4))

	gl.BindVertexArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	if t.content != "" {
		t.generateGeometry()
	}

	return nil
}

func (t *Text) generateGeometry() {
	if t.content == "" {
		t.vertexCount = 0
		return
	}

	vertices := make([]float32, 0)
	cursorX := t.x

	for _, ch := range t.content {
		glyph, ok := t.font.Glyphs[ch]
		if !ok {
			continue // Skip unknown characters
		}

		xpos := cursorX + glyph.Bearing.X()*t.scale
		ypos := t.y + (glyph.Bearing.Y()-glyph.Size.Y())*t.scale

		w := glyph.Size.X() * t.scale
		h := glyph.Size.Y() * t.scale

		// Append Quad (2 Triangles)
		// V1 (Top-Left)
		vertices = append(vertices, xpos, ypos, t.color.X(), t.color.Y(), t.color.Z(), glyph.UVMin.X(), glyph.UVMin.Y())
		// V2 (Top-Right)
		vertices = append(vertices, xpos+w, ypos, t.color.X(), t.color.Y(), t.color.Z(), glyph.UVMax.X(), glyph.UVMin.Y())
		// V3 (Bottom-Right)
		vertices = append(vertices, xpos+w, ypos+h, t.color.X(), t.color.Y(), t.color.Z(), glyph.UVMax.X(), glyph.UVMax.Y())
		// V4 (Top-Left again)
		vertices = append(vertices, xpos, ypos, t.color.X(), t.color.Y(), t.color.Z(), glyph.UVMin.X(), glyph.UVMin.Y())
		// V5 (Bottom-Right again)
		vertices = append(vertices, xpos+w, ypos+h, t.color.X(), t.color.Y(), t.color.Z(), glyph.UVMax.X(), glyph.UVMax.Y())
		// V6 (Bottom-Left)
		vertices = append(vertices, xpos, ypos+h, t.color.X(), t.color.Y(), t.color.Z(), glyph.UVMin.X(), glyph.UVMax.Y())

		// Move cursor for next character
		cursorX += glyph.Advance * t.scale
	}

	t.vertexCount = int32(len(vertices) / 7)

	gl.BindBuffer(gl.ARRAY_BUFFER, t.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.DYNAMIC_DRAW)

	t.needsUpdate = false
}

func (t *Text) SetContent(content string) {
	if t.content != content {
		t.content = content
		t.needsUpdate = true
	}
}

func (t *Text) Update(state interface{}) {
	if t.needsUpdate {
		t.generateGeometry()
	}
}

func (t *Text) Draw(shaderProgram uint32, projection mgl32.Mat4) {
	if t.vertexCount <= 0 {
		return
	}

	// Bind Font Texture (Override the default white texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.font.TextureID)

	// Draw Text
	gl.BindVertexArray(t.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, t.vertexCount)
	gl.BindVertexArray(0)
}

func (t *Text) Cleanup() {
	gl.DeleteVertexArrays(1, &t.vao)
	gl.DeleteBuffers(1, &t.vbo)
}
