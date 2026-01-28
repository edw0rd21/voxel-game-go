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
var chunkVertexSource string

//go:embed shaders/fragment.glsl
var chunkFragmentSource string

//go:embed shaders/flat_vertex.glsl
var flatVertexSource string

//go:embed shaders/flat_fragment.glsl
var flatFragmentSource string

type Renderer struct {
	// Program 1: The World (Chunks)
	chunkProgram       uint32
	chunkUniModel      int32
	chunkUniView       int32
	chunkUniProjection int32

	// Program 2: Flat Geometry (Highlights)
	flatProgram       uint32
	flatUniModel      int32
	flatUniView       int32
	flatUniProjection int32
	flatUniColor      int32

	highlightVAO uint32
	highlightVBO uint32
}

func NewRenderer() (*Renderer, error) {
	renderer := &Renderer{}

	// Compile Chunk Shader
	var err error
	renderer.chunkProgram, err = createProgram(chunkVertexSource, chunkFragmentSource)
	if err != nil {
		return nil, fmt.Errorf("failed to create chunk shader: %v", err)
	}

	// Compile Flat Shader
	renderer.flatProgram, err = createProgram(flatVertexSource, flatFragmentSource)
	if err != nil {
		return nil, fmt.Errorf("failed to create flat shader: %v", err)
	}

	// Cache Uniforms for Chunk Program
	gl.UseProgram(renderer.chunkProgram)
	renderer.chunkUniModel = gl.GetUniformLocation(renderer.chunkProgram, gl.Str("model\x00"))
	renderer.chunkUniView = gl.GetUniformLocation(renderer.chunkProgram, gl.Str("view\x00"))
	renderer.chunkUniProjection = gl.GetUniformLocation(renderer.chunkProgram, gl.Str("projection\x00"))

	// Cache Uniforms for Flat Program
	gl.UseProgram(renderer.flatProgram)
	renderer.flatUniModel = gl.GetUniformLocation(renderer.flatProgram, gl.Str("model\x00"))
	renderer.flatUniView = gl.GetUniformLocation(renderer.flatProgram, gl.Str("view\x00"))
	renderer.flatUniProjection = gl.GetUniformLocation(renderer.flatProgram, gl.Str("projection\x00"))
	renderer.flatUniColor = gl.GetUniformLocation(renderer.flatProgram, gl.Str("uColor\x00"))

	gl.UseProgram(0)

	renderer.initHighlightMesh()
	return renderer, nil
}

func (r *Renderer) RenderWorld(w *world.World, cam *camera.Camera) {
	gl.UseProgram(r.chunkProgram)

	// Set view and projection matrices
	view := cam.GetViewMatrix()
	projection := cam.GetProjectionMatrix()

	gl.UniformMatrix4fv(r.chunkUniView, 1, false, &view[0])
	gl.UniformMatrix4fv(r.chunkUniProjection, 1, false, &projection[0])

	// Render each chunk
	//chunksRendered := 0
	for _, chunk := range w.GetChunks() {
		if chunk.Mesh == nil || chunk.Mesh.VertexCount == 0 {
			continue
		}

		// Frustum culling
		// if !cam.IsChunkVisible(chunk.X, chunk.Z, 16) {
		// 	continue
		// }

		// Set model matrix (identity for now, chunk position handled in vertex data)
		model := mgl32.Ident4()
		gl.UniformMatrix4fv(r.chunkUniModel, 1, false, &model[0])

		// Bind and draw
		gl.BindVertexArray(chunk.Mesh.VAO)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(chunk.Mesh.VertexCount))
		gl.BindVertexArray(0)

		//chunksRendered++
	}
}

func (r *Renderer) DrawBlockHighlight(pos mgl32.Vec3, cam *camera.Camera, color mgl32.Vec3) {
	gl.UseProgram(r.flatProgram)

	model := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z()).
		Mul4(mgl32.Scale3D(1.001, 1.001, 1.001))

	view := cam.GetViewMatrix()
	proj := cam.GetProjectionMatrix()

	// Use cached uniforms
	gl.UniformMatrix4fv(r.flatUniModel, 1, false, &model[0])
	gl.UniformMatrix4fv(r.flatUniView, 1, false, &view[0])
	gl.UniformMatrix4fv(r.flatUniProjection, 1, false, &proj[0])
	gl.Uniform3fv(r.flatUniColor, 1, &color[0])

	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)
	gl.Disable(gl.CULL_FACE)

	gl.BindVertexArray(r.highlightVAO)

	// 12 beams * 36 vertices per beam = 432 vertices
	gl.DrawArrays(gl.TRIANGLES, 0, 432)

	gl.BindVertexArray(0)

	gl.DepthMask(true)
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
}

func (r *Renderer) initHighlightMesh() {
	var vertices []float32
	thickness := float32(0.02) // Adjust for thicker/thinner lines

	// Helper to create a box (cube) at specific pos with specific size
	addBeam := func(x, y, z, w, h, d float32) {
		// Min/Max bounds for this beam
		x1, y1, z1 := x, y, z
		x2, y2, z2 := x+w, y+h, z+d

		cube := []float32{
			// Front face
			x1, y1, z2, x2, y1, z2, x2, y2, z2,
			x2, y2, z2, x1, y2, z2, x1, y1, z2,
			// Back face
			x2, y1, z1, x1, y1, z1, x1, y2, z1,
			x1, y2, z1, x2, y2, z1, x2, y1, z1,
			// Left face
			x1, y1, z1, x1, y1, z2, x1, y2, z2,
			x1, y2, z2, x1, y2, z1, x1, y1, z1,
			// Right face
			x2, y1, z2, x2, y1, z1, x2, y2, z1,
			x2, y2, z1, x2, y2, z2, x2, y1, z2,
			// Top face
			x1, y2, z2, x2, y2, z2, x2, y2, z1,
			x2, y2, z1, x1, y2, z1, x1, y2, z2,
			// Bottom face
			x1, y1, z1, x2, y1, z1, x2, y1, z2,
			x2, y1, z2, x1, y1, z2, x1, y1, z1,
		}
		vertices = append(vertices, cube...)
	}

	// Generate the 12 edges
	// Vertical Edges (4)
	addBeam(0, 0, 0, thickness, 1, thickness)                     // Front-Left
	addBeam(1-thickness, 0, 0, thickness, 1, thickness)           // Front-Right
	addBeam(1-thickness, 0, 1-thickness, thickness, 1, thickness) // Back-Right
	addBeam(0, 0, 1-thickness, thickness, 1, thickness)           // Back-Left

	// Top Horizontal Edges (4)
	addBeam(0, 1-thickness, 0, 1, thickness, thickness)           // Front
	addBeam(0, 1-thickness, 1-thickness, 1, thickness, thickness) // Back
	addBeam(0, 1-thickness, 0, thickness, thickness, 1)           // Left
	addBeam(1-thickness, 1-thickness, 0, thickness, thickness, 1) // Right

	// Bottom Horizontal Edges (4)
	addBeam(0, 0, 0, 1, thickness, thickness)           // Front
	addBeam(0, 0, 1-thickness, 1, thickness, thickness) // Back
	addBeam(0, 0, 0, thickness, thickness, 1)           // Left
	addBeam(1-thickness, 0, 0, thickness, thickness, 1) // Right

	gl.GenVertexArrays(1, &r.highlightVAO)
	gl.GenBuffers(1, &r.highlightVBO)

	gl.BindVertexArray(r.highlightVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.highlightVBO)

	gl.BufferData(
		gl.ARRAY_BUFFER,
		len(vertices)*4,
		gl.Ptr(vertices),
		gl.STATIC_DRAW,
	)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(
		0,
		3,
		gl.FLOAT,
		false,
		3*4,
		gl.PtrOffset(0),
	)

	gl.BindVertexArray(0)
}

func createProgram(vertexSource, fragmentSource string) (uint32, error) {
	vertexShader, err := compileShader(vertexSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	fragmentShader, err := compileShader(fragmentSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
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
