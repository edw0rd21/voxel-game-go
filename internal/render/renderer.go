package render

import (
	"fmt"
	"os"
	"strings"

	"voxel-game/internal/camera"
	"voxel-game/internal/world"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type Renderer struct {
	shaderProgram   uint32
	highlightShader uint32 // For block selection
	highlightVAO    uint32
	highlightVBO    uint32
}

func NewRenderer() (*Renderer, error) {
	// Compile Main Shader
	shaderProgram, err := createShaderProgram("internal/render/shaders/vertex.glsl", "internal/render/shaders/fragment.glsl")
	if err != nil {
		return nil, fmt.Errorf("failed to create main shader: %w", err)
	}

	// Compile Highlight Shader
	highlightShader, err := createShaderProgram("internal/render/shaders/flat_vertex.glsl", "internal/render/shaders/flat_fragment.glsl")
	if err != nil {
		// Fallback or error handling
		fmt.Println("Warning: Could not load highlight shader, using main:", err)
		highlightShader = shaderProgram
	}

	r := &Renderer{
		shaderProgram:   shaderProgram,
		highlightShader: highlightShader,
	}
	r.initHighlightMesh()

	return r, nil
}

func (r *Renderer) RenderWorld(w *world.World, cam *camera.Camera, atlasTextureID uint32) {
	gl.UseProgram(r.shaderProgram)

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, atlasTextureID)

	loc := gl.GetUniformLocation(r.shaderProgram, gl.Str("texture1\x00"))
	gl.Uniform1i(loc, 0)

	// Set view and projection matrices
	view := cam.GetViewMatrix()
	projection := cam.GetProjectionMatrix()

	viewLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("view\x00"))
	projLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("projection\x00"))
	modelLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("model\x00"))
	lightLoc := gl.GetUniformLocation(r.shaderProgram, gl.Str("lightDir\x00"))

	gl.UniformMatrix4fv(viewLoc, 1, false, &view[0])
	gl.UniformMatrix4fv(projLoc, 1, false, &projection[0])

	// Simple directional light
	lightDir := mgl32.Vec3{-0.2, -1.0, -0.3}
	gl.Uniform3fv(lightLoc, 1, &lightDir[0])

	// Render each chunk
	for _, chunk := range w.GetChunks() {
		if chunk.Mesh == nil || chunk.Mesh.VertexCount == 0 {
			continue
		}

		// Frustum culling
		if !cam.IsChunkVisible(chunk.X, chunk.Z, world.ChunkSize) {
			continue
		}

		// Set model matrix
		model := mgl32.Ident4()
		gl.UniformMatrix4fv(modelLoc, 1, false, &model[0])

		gl.BindVertexArray(chunk.Mesh.VAO)
		gl.DrawArrays(gl.TRIANGLES, 0, int32(chunk.Mesh.VertexCount))
	}
	gl.BindVertexArray(0)
}

func (r *Renderer) DrawBlockHighlight(pos mgl32.Vec3, cam *camera.Camera, color mgl32.Vec3) {
	gl.UseProgram(r.highlightShader)

	model := mgl32.Translate3D(pos.X(), pos.Y(), pos.Z()).
		Mul4(mgl32.Scale3D(1.001, 1.001, 1.001))

	view := cam.GetViewMatrix()
	proj := cam.GetProjectionMatrix()

	// Use cached uniforms
	gl.UniformMatrix4fv(gl.GetUniformLocation(r.highlightShader, gl.Str("model\x00")), 1, false, &model[0])
	gl.UniformMatrix4fv(gl.GetUniformLocation(r.highlightShader, gl.Str("view\x00")), 1, false, &view[0])
	gl.UniformMatrix4fv(gl.GetUniformLocation(r.highlightShader, gl.Str("projection\x00")), 1, false, &proj[0])
	gl.Uniform3fv(gl.GetUniformLocation(r.highlightShader, gl.Str("color\x00")), 1, &color[0])

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

func createShaderProgram(vertexPath, fragmentPath string) (uint32, error) {
	vertexSource, err := os.ReadFile(vertexPath)
	if err != nil {
		return 0, err
	}
	fragmentSource, err := os.ReadFile(fragmentPath)
	if err != nil {
		return 0, err
	}

	vertexShader, err := compileShader(string(vertexSource)+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	fragmentShader, err := compileShader(string(fragmentSource)+"\x00", gl.FRAGMENT_SHADER)
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

	csources, free := gl.Strs(source)
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
