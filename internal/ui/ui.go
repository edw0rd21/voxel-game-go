package ui

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

//go:embed shaders/ui_vertex.glsl
var uiVertexShaderSource string

//go:embed shaders/ui_fragment.glsl
var uiFragmentShaderSource string

// UIElement interface - all UI elements must implement this
type UIElement interface {
	Init() error
	Update(state interface{})
	Draw(shaderProgram uint32, projection mgl32.Mat4)
	Cleanup()
}

// UIRenderer manages all UI elements and orchestrates rendering
type UIRenderer struct {
	shaderProgram uint32
	projection    mgl32.Mat4
	elements      []UIElement
	width         int
	height        int
}

func NewUIRenderer(width, height int) (*UIRenderer, error) {
	// Compile shaders
	vertexShader, err := compileShader(uiVertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}

	fragmentShader, err := compileShader(uiFragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}

	// Link shader program
	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	// Check linking status
	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to link UI shader program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	// Create orthographic projection matrix
	projection := mgl32.Ortho(0, float32(width), float32(height), 0, -1, 1)

	renderer := &UIRenderer{
		shaderProgram: program,
		projection:    projection,
		elements:      make([]UIElement, 0),
		width:         width,
		height:        height,
	}

	return renderer, nil
}

func (r *UIRenderer) AddElement(element UIElement) error {
	if err := element.Init(); err != nil {
		return err
	}
	r.elements = append(r.elements, element)
	return nil
}

func (r *UIRenderer) Resize(width, height int) {
	r.width = width
	r.height = height
	r.projection = mgl32.Ortho(0, float32(width), float32(height), 0, -1, 1)
}

func (r *UIRenderer) Render() {
	// Disable depth test for UI overlay
	gl.Disable(gl.DEPTH_TEST)

	gl.UseProgram(r.shaderProgram)

	// Set projection matrix
	projectionLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("projection\x00"))
	gl.UniformMatrix4fv(projectionLoc, 1, false, &r.projection[0])

	// Draw all elements in order (determines layering)
	for _, element := range r.elements {
		element.Draw(r.shaderProgram, r.projection)
	}

	// Re-enable depth test
	gl.Enable(gl.DEPTH_TEST)
}

func (r *UIRenderer) Cleanup() {
	for _, element := range r.elements {
		element.Cleanup()
	}
	gl.DeleteProgram(r.shaderProgram)
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile UI shader: %v", log)
	}

	return shader, nil
}
