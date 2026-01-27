package world

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/ojrac/opensimplex-go"
)

const (
	ChunkSize      = 16
	ChunkHeight    = 64
	RenderDistance = 16
)

type BlockType uint8

const (
	BlockAir BlockType = iota
	BlockDirt
	BlockGrass
	BlockStone
)

type Block struct {
	Type BlockType
}

type Chunk struct {
	X, Z   int
	Blocks [ChunkSize][ChunkHeight][ChunkSize]Block
	Mesh   *ChunkMesh
}

type ChunkMesh struct {
	VAO         uint32
	VBO         uint32
	VertexCount int
}

type World struct {
	chunks map[[2]int]*Chunk
	noise  opensimplex.Noise
}

func NewWorld() *World {
	w := &World{
		chunks: make(map[[2]int]*Chunk),
		noise:  opensimplex.NewNormalized(12345),
	}

	// Generate initial chunks around spawn
	for x := -2; x <= 2; x++ {
		for z := -2; z <= 2; z++ {
			chunk := w.generateChunk(x, z)
			w.chunks[[2]int{x, z}] = chunk
			chunk.generateMesh(w)
		}
	}

	return w
}

func (w *World) generateChunk(chunkX, chunkZ int) *Chunk {
	chunk := &Chunk{
		X: chunkX,
		Z: chunkZ,
	}

	// Multi-octave terrain generation
	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {
			worldX := float64(chunkX*ChunkSize + x)
			worldZ := float64(chunkZ*ChunkSize + z)

			// Layer 1: Base terrain (large rolling hills)
			continentalness := w.noise.Eval2(worldX*0.005, worldZ*0.005)

			// Layer 2: Medium features (hills and valleys)
			erosion := w.noise.Eval2(worldX*0.02, worldZ*0.02)

			// Layer 3: Fine detail (surface variation)
			detail := w.noise.Eval2(worldX*0.1, worldZ*0.1)

			// Combine the layers with different weights
			// Noise values range from -1 to 1, normalize to height
			baseHeight := 35.0
			continentalScale := 25.0
			erosionScale := 10.0
			detailScale := 3.0

			height := baseHeight +
				continentalness*continentalScale +
				erosion*erosionScale +
				detail*detailScale

			heightInt := int(height)

			// Generate terrain layers
			for y := 0; y < ChunkHeight; y++ {
				if y < heightInt-4 {
					// Deep underground = stone
					chunk.Blocks[x][y][z].Type = BlockStone
				} else if y < heightInt {
					// Near surface = dirt
					chunk.Blocks[x][y][z].Type = BlockDirt
				} else if y == heightInt {
					// Surface = grass
					chunk.Blocks[x][y][z].Type = BlockGrass
				} else {
					// Above ground = air
					chunk.Blocks[x][y][z].Type = BlockAir
				}
			}
		}
	}

	return chunk
}

func (c *Chunk) generateMesh(w *World) {
	var vertices []float32

	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkHeight; y++ {
			for z := 0; z < ChunkSize; z++ {
				block := c.Blocks[x][y][z]
				if block.Type == BlockAir {
					continue
				}

				worldX := float32(c.X*ChunkSize + x)
				worldY := float32(y)
				worldZ := float32(c.Z*ChunkSize + z)

				color := getBlockColor(block.Type)

				// Check each face and add if exposed
				// Front face (+Z)
				if c.isTransparent(w, x, y, z+1) {
					vertices = append(vertices, createFace(worldX, worldY, worldZ, 0, color)...)
				}
				// Back face (-Z)
				if c.isTransparent(w, x, y, z-1) {
					vertices = append(vertices, createFace(worldX, worldY, worldZ, 1, color)...)
				}
				// Right face (+X)
				if c.isTransparent(w, x+1, y, z) {
					vertices = append(vertices, createFace(worldX, worldY, worldZ, 2, color)...)
				}
				// Left face (-X)
				if c.isTransparent(w, x-1, y, z) {
					vertices = append(vertices, createFace(worldX, worldY, worldZ, 3, color)...)
				}
				// Top face (+Y)
				if c.isTransparent(w, x, y+1, z) {
					vertices = append(vertices, createFace(worldX, worldY, worldZ, 4, color)...)
				}
				// Bottom face (-Y)
				if c.isTransparent(w, x, y-1, z) {
					vertices = append(vertices, createFace(worldX, worldY, worldZ, 5, color)...)
				}
			}
		}
	}

	if len(vertices) == 0 {
		return
	}

	// Create mesh
	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Color attribute
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)

	c.Mesh = &ChunkMesh{
		VAO:         vao,
		VBO:         vbo,
		VertexCount: len(vertices) / 6,
	}
}

func (c *Chunk) isTransparent(w *World, x, y, z int) bool {
	// Check bounds
	if y < 0 || y >= ChunkHeight {
		return true
	}

	// If within this chunk
	if x >= 0 && x < ChunkSize && z >= 0 && z < ChunkSize {
		return c.Blocks[x][y][z].Type == BlockAir
	}

	// Check neighboring chunk
	chunkX := c.X
	chunkZ := c.Z
	localX := x
	localZ := z

	if x < 0 {
		chunkX--
		localX = ChunkSize - 1
	} else if x >= ChunkSize {
		chunkX++
		localX = 0
	}

	if z < 0 {
		chunkZ--
		localZ = ChunkSize - 1
	} else if z >= ChunkSize {
		chunkZ++
		localZ = 0
	}

	neighbor, exists := w.chunks[[2]int{chunkX, chunkZ}]
	if !exists {
		return true
	}

	return neighbor.Blocks[localX][y][localZ].Type == BlockAir
}

func createFace(x, y, z float32, face int, color mgl32.Vec3) []float32 {
	// Each face is 2 triangles = 6 vertices
	// Each vertex is 6 floats: 3 for position, 3 for color
	vertices := make([]float32, 36)

	var positions [6][3]float32

	switch face {
	case 0: // Front (+Z)
		positions = [6][3]float32{
			{x, y, z + 1}, {x + 1, y, z + 1}, {x + 1, y + 1, z + 1},
			{x, y, z + 1}, {x + 1, y + 1, z + 1}, {x, y + 1, z + 1},
		}
	case 1: // Back (-Z)
		positions = [6][3]float32{
			{x + 1, y, z}, {x, y, z}, {x, y + 1, z},
			{x + 1, y, z}, {x, y + 1, z}, {x + 1, y + 1, z},
		}
	case 2: // Right (+X)
		positions = [6][3]float32{
			{x + 1, y, z + 1}, {x + 1, y, z}, {x + 1, y + 1, z},
			{x + 1, y, z + 1}, {x + 1, y + 1, z}, {x + 1, y + 1, z + 1},
		}
	case 3: // Left (-X)
		positions = [6][3]float32{
			{x, y, z}, {x, y, z + 1}, {x, y + 1, z + 1},
			{x, y, z}, {x, y + 1, z + 1}, {x, y + 1, z},
		}
	case 4: // Top (+Y)
		positions = [6][3]float32{
			{x, y + 1, z + 1}, {x + 1, y + 1, z + 1}, {x + 1, y + 1, z},
			{x, y + 1, z + 1}, {x + 1, y + 1, z}, {x, y + 1, z},
		}
	case 5: // Bottom (-Y)
		positions = [6][3]float32{
			{x, y, z}, {x + 1, y, z}, {x + 1, y, z + 1},
			{x, y, z}, {x + 1, y, z + 1}, {x, y, z + 1},
		}
	}

	// Add slight shading based on face direction
	shadingFactor := float32(1.0)
	switch face {
	case 0: // Front (+Z)
		shadingFactor = 0.85 // CHANGE from 1.0
	case 1: // Back (-Z)
		shadingFactor = 0.75 // CHANGE from 0.8
	case 2, 3: // Sides (X)
		shadingFactor = 0.80 // CHANGE from 0.9
	case 4: // Top (+Y)
		shadingFactor = 1.0 // Keep brightest
	case 5: // Bottom (-Y)
		shadingFactor = 0.6 // CHANGE from 0.7
	}

	shadedColor := color.Mul(shadingFactor)

	for i := 0; i < 6; i++ {
		vertices[i*6+0] = positions[i][0]
		vertices[i*6+1] = positions[i][1]
		vertices[i*6+2] = positions[i][2]
		vertices[i*6+3] = shadedColor[0]
		vertices[i*6+4] = shadedColor[1]
		vertices[i*6+5] = shadedColor[2]
	}

	return vertices
}

func getBlockColor(blockType BlockType) mgl32.Vec3 {
	switch blockType {
	case BlockGrass:
		return mgl32.Vec3{0.2, 0.8, 0.2} // Green
	case BlockDirt:
		return mgl32.Vec3{0.6, 0.4, 0.2} // Brown
	case BlockStone:
		return mgl32.Vec3{0.5, 0.5, 0.5} // Gray
	default:
		return mgl32.Vec3{1, 1, 1} // White
	}
}

func (w *World) GetChunks() []*Chunk {
	chunks := make([]*Chunk, 0, len(w.chunks))
	for _, chunk := range w.chunks {
		chunks = append(chunks, chunk)
	}
	return chunks
}

func (w *World) GetBlock(x, y, z int) BlockType {
	if y < 0 || y >= ChunkHeight {
		return BlockAir
	}

	chunkX := x / ChunkSize
	chunkZ := z / ChunkSize
	localX := x % ChunkSize
	localZ := z % ChunkSize

	if localX < 0 {
		localX += ChunkSize
		chunkX--
	}
	if localZ < 0 {
		localZ += ChunkSize
		chunkZ--
	}

	chunk, exists := w.chunks[[2]int{chunkX, chunkZ}]
	if !exists {
		return BlockAir
	}

	return chunk.Blocks[localX][y][localZ].Type
}

func (w *World) SetBlock(x, y, z int, blockType BlockType) {
	if y < 0 || y >= ChunkHeight {
		return
	}

	chunkX := x / ChunkSize
	chunkZ := z / ChunkSize
	localX := x % ChunkSize
	localZ := z % ChunkSize

	if localX < 0 {
		localX += ChunkSize
		chunkX--
	}
	if localZ < 0 {
		localZ += ChunkSize
		chunkZ--
	}

	chunk, exists := w.chunks[[2]int{chunkX, chunkZ}]
	if !exists {
		return
	}

	chunk.Blocks[localX][y][localZ].Type = blockType

	// Regenerate mesh
	chunk.generateMesh(w)

	// Also regenerate neighboring chunks if block is on edge
	if localX == 0 {
		if neighbor, ok := w.chunks[[2]int{chunkX - 1, chunkZ}]; ok {
			neighbor.generateMesh(w)
		}
	} else if localX == ChunkSize-1 {
		if neighbor, ok := w.chunks[[2]int{chunkX + 1, chunkZ}]; ok {
			neighbor.generateMesh(w)
		}
	}

	if localZ == 0 {
		if neighbor, ok := w.chunks[[2]int{chunkX, chunkZ - 1}]; ok {
			neighbor.generateMesh(w)
		}
	} else if localZ == ChunkSize-1 {
		if neighbor, ok := w.chunks[[2]int{chunkX, chunkZ + 1}]; ok {
			neighbor.generateMesh(w)
		}
	}
}

func (w *World) UpdateChunks(playerX, playerZ float32) {
	// Calculate which chunk the player is in
	playerChunkX := int(math.Floor(float64(playerX))) / ChunkSize
	playerChunkZ := int(math.Floor(float64(playerZ))) / ChunkSize

	// Generate chunks in render distance
	for x := playerChunkX - RenderDistance; x <= playerChunkX+RenderDistance; x++ {
		for z := playerChunkZ - RenderDistance; z <= playerChunkZ+RenderDistance; z++ {
			chunkKey := [2]int{x, z}

			// If chunk doesn't exist, generate it
			if _, exists := w.chunks[chunkKey]; !exists {
				chunk := w.generateChunk(x, z)
				w.chunks[chunkKey] = chunk
				chunk.generateMesh(w)
			}
		}
	}

	// Unload chunks that are too far away
	toDelete := make([][2]int, 0)
	for key := range w.chunks {
		dx := key[0] - playerChunkX
		dz := key[1] - playerChunkZ
		distance := math.Sqrt(float64(dx*dx + dz*dz))

		if distance > float64(RenderDistance+2) {
			// Clean up OpenGL resources
			if w.chunks[key].Mesh != nil {
				gl.DeleteVertexArrays(1, &w.chunks[key].Mesh.VAO)
				gl.DeleteBuffers(1, &w.chunks[key].Mesh.VBO)
			}
			toDelete = append(toDelete, key)
		}
	}

	// Delete the chunks
	for _, key := range toDelete {
		delete(w.chunks, key)
	}
}
