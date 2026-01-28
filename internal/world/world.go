package world

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
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
	vertices := make([]float32, 0, 4096)

	// Cache neighbors to avoid map lookups in the inner loop
	nLeft := w.chunks[[2]int{c.X - 1, c.Z}]
	nRight := w.chunks[[2]int{c.X + 1, c.Z}]
	nBack := w.chunks[[2]int{c.X, c.Z - 1}]
	nFront := w.chunks[[2]int{c.X, c.Z + 1}]

	// Helper closure to check transparency quickly
	isTransparent := func(x, y, z int) bool {
		if y < 0 || y >= ChunkHeight {
			return true
		}

		// Internal check (Fastest)
		if x >= 0 && x < ChunkSize && z >= 0 && z < ChunkSize {
			return c.Blocks[x][y][z].Type == BlockAir
		}

		// Neighbor checks (Fast-ish, using cached pointers)
		if x < 0 {
			if nLeft == nil {
				return true
			}
			return nLeft.Blocks[ChunkSize-1][y][z].Type == BlockAir
		}
		if x >= ChunkSize {
			if nRight == nil {
				return true
			}
			return nRight.Blocks[0][y][z].Type == BlockAir
		}
		if z < 0 {
			if nBack == nil {
				return true
			}
			return nBack.Blocks[x][y][ChunkSize-1].Type == BlockAir
		}
		if z >= ChunkSize {
			if nFront == nil {
				return true
			}
			return nFront.Blocks[x][y][0].Type == BlockAir
		}
		return true
	}

	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkHeight; y++ {
			for z := 0; z < ChunkSize; z++ {
				blockType := c.Blocks[x][y][z].Type
				if blockType == BlockAir {
					continue
				}

				worldX := float32(c.X*ChunkSize + x)
				worldY := float32(y)
				worldZ := float32(c.Z*ChunkSize + z)

				// Get base color
				r, g, b := getBlockColorRGB(blockType)

				// INLINED FACE GENERATION
				// Front face (+Z)
				if isTransparent(x, y, z+1) {
					addFace(&vertices, worldX, worldY, worldZ, 0, r, g, b)
				}
				// Back face (-Z)
				if isTransparent(x, y, z-1) {
					addFace(&vertices, worldX, worldY, worldZ, 1, r, g, b)
				}
				// Right face (+X)
				if isTransparent(x+1, y, z) {
					addFace(&vertices, worldX, worldY, worldZ, 2, r, g, b)
				}
				// Left face (-X)
				if isTransparent(x-1, y, z) {
					addFace(&vertices, worldX, worldY, worldZ, 3, r, g, b)
				}
				// Top face (+Y)
				if isTransparent(x, y+1, z) {
					addFace(&vertices, worldX, worldY, worldZ, 4, r, g, b)
				}
				// Bottom face (-Y)
				if isTransparent(x, y-1, z) {
					addFace(&vertices, worldX, worldY, worldZ, 5, r, g, b)
				}
			}
		}
	}

	if len(vertices) == 0 {
		return
	}

	// Create mesh
	if c.Mesh == nil {
		c.Mesh = &ChunkMesh{}
		gl.GenVertexArrays(1, &c.Mesh.VAO)
		gl.GenBuffers(1, &c.Mesh.VBO)
	}

	gl.BindVertexArray(c.Mesh.VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, c.Mesh.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Position attribute
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Color attribute
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindVertexArray(0)
	c.Mesh.VertexCount = len(vertices) / 6
}

// Return floats directly to avoid Vec3 allocation
func getBlockColorRGB(blockType BlockType) (float32, float32, float32) {
	switch blockType {
	case BlockGrass:
		return 0.2, 0.8, 0.2
	case BlockDirt:
		return 0.6, 0.4, 0.2
	case BlockStone:
		return 0.5, 0.5, 0.5
	default:
		return 1.0, 1.0, 1.0
	}
}

// Helper to append vertices directly
func addFace(verts *[]float32, x, y, z float32, face int, r, g, b float32) {
	// Apply shading
	shade := float32(1.0)
	switch face {
	case 0:
		shade = 0.85 // Front
	case 1:
		shade = 0.75 // Back
	case 2, 3:
		shade = 0.80 // Sides
	case 4:
		shade = 1.0 // Top
	case 5:
		shade = 0.6 // Bottom
	}

	r *= shade
	g *= shade
	b *= shade

	// Append 6 vertices (2 triangles)
	// Front (+Z)
	if face == 0 {
		*verts = append(*verts,
			x, y, z+1, r, g, b,
			x+1, y, z+1, r, g, b,
			x+1, y+1, z+1, r, g, b,
			x, y, z+1, r, g, b,
			x+1, y+1, z+1, r, g, b,
			x, y+1, z+1, r, g, b,
		)
	} else if face == 1 { // Back (-Z)
		*verts = append(*verts,
			x+1, y, z, r, g, b,
			x, y, z, r, g, b,
			x, y+1, z, r, g, b,
			x+1, y, z, r, g, b,
			x, y+1, z, r, g, b,
			x+1, y+1, z, r, g, b,
		)
	} else if face == 2 { // Right (+X)
		*verts = append(*verts,
			x+1, y, z+1, r, g, b,
			x+1, y, z, r, g, b,
			x+1, y+1, z, r, g, b,
			x+1, y, z+1, r, g, b,
			x+1, y+1, z, r, g, b,
			x+1, y+1, z+1, r, g, b,
		)
	} else if face == 3 { // Left (-X)
		*verts = append(*verts,
			x, y, z, r, g, b,
			x, y, z+1, r, g, b,
			x, y+1, z+1, r, g, b,
			x, y, z, r, g, b,
			x, y+1, z+1, r, g, b,
			x, y+1, z, r, g, b,
		)
	} else if face == 4 { // Top (+Y)
		*verts = append(*verts,
			x, y+1, z+1, r, g, b,
			x+1, y+1, z+1, r, g, b,
			x+1, y+1, z, r, g, b,
			x, y+1, z+1, r, g, b,
			x+1, y+1, z, r, g, b,
			x, y+1, z, r, g, b,
		)
	} else if face == 5 { // Bottom (-Y)
		*verts = append(*verts,
			x, y, z, r, g, b,
			x+1, y, z, r, g, b,
			x+1, y, z+1, r, g, b,
			x, y, z, r, g, b,
			x+1, y, z+1, r, g, b,
			x, y, z+1, r, g, b,
		)
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
