package render

import (
	_ "embed"
	"fmt"
	"strings"

	"voxel-game/internal/camera"
	"voxel-game/internal/world"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

//go:embed shaders/vertex.glsl
var vertexShaderSource string

//go:embed shaders/fragment.glsl
var fragmentShaderSource string

type Renderer struct {
	shaderProgram uint32
	crosshairVAO  uint32
	crosshairVBO  uint32
}

func NewRenderer() (*Renderer, error) {
	// Compile shaders
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
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
		return nil, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return &Renderer{
		shaderProgram: program,
	}, nil
}

func (r *Renderer) RenderWorld(w *world.World, cam *camera.Camera) {
	gl.UseProgram(r.shaderProgram)

	// Get uniform locations
	modelLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("model\x00"))
	viewLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("view\x00"))
	projectionLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("projection\x00"))

	// Set view and projection matrices
	view := cam.GetViewMatrix()
	projection := cam.GetProjectionMatrix()

	gl.UniformMatrix4fv(viewLoc, 1, false, &view[0])
	gl.UniformMatrix4fv(projectionLoc, 1, false, &projection[0])

	// Render each chunk
	chunksRendered := 0
	for _, chunk := range w.GetChunks() {
		if chunk.Mesh == nil || chunk.Mesh.VertexCount == 0 {
			continue
		}

		// Frustum culling
		if !cam.IsChunkVisible(chunk.X, chunk.Z, 16) {
			continue
		}

		// Set model matrix (identity for now, chunk position handled in vertex data)
		model := mgl32.Ident4()
		gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])

		// Bind and draw
		gl.BindVertexArray(chunk.Mesh.VAO)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(chunk.Mesh.VertexCount))
		gl.BindVertexArray(0)

		chunksRendered++
	}
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

		return 0, fmt.Errorf("failed to compile shader: %v", log)
	}

	return shader, nil
}
