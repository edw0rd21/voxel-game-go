package world

import (
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/ojrac/opensimplex-go"
)

const (
	ChunkSize      = 16
	ChunkHeight    = 256
	RenderDistance = 16
)

type Block struct {
	Type BlockType
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

	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {

			absX := chunkX*ChunkSize + x
			absZ := chunkZ*ChunkSize + z

			worldX := float64(absX)
			worldZ := float64(absZ)

			// Terrain Gen
			ruggedness := w.noise.Eval2(worldX*0.004, worldZ*0.004)
			jitter := w.noise.Eval2(worldX*0.5, worldZ*0.5)

			mountainShape := math.Abs(w.noise.Eval2(worldX*0.015, worldZ*0.015))
			mountainShape = math.Pow(mountainShape, 2)

			baseElevation := w.noise.Eval2(worldX*0.005, worldZ*0.005)
			var amplitude float64

			if ruggedness > 0.6 {
				factor := math.Min((ruggedness-0.6)/0.4, 1.0)
				amplitude = 40.0 + (factor * 100.0)
			} else if ruggedness > 0.2 {
				factor := math.Min((ruggedness-0.2)/0.4, 1.0)
				amplitude = 10.0 + (factor * 30.0)
			} else {
				amplitude = 2.0 + ((ruggedness + 1.0) * 8.0)
			}

			baseLevel := 25.0
			height := baseLevel +
				(baseElevation * 20.0) +
				(mountainShape * amplitude) +
				(w.noise.Eval2(worldX*0.1, worldZ*0.1) * 2.0)

			if height < 2 {
				height = 2
			}
			if height > ChunkHeight-5 {
				height = ChunkHeight - 5
			}

			heightInt := int(height)

			for y := 0; y < ChunkHeight; y++ {
				if y == 0 {
					chunk.Blocks[x][y][z].Type = BlockStone
					continue
				}
				if y > heightInt {
					chunk.Blocks[x][y][z].Type = BlockAir
					continue
				}

				if y == heightInt {
					if y > 90+int(jitter*11) {
						chunk.Blocks[x][y][z].Type = BlockSnow
					} else if y > 72+int(jitter*7) {
						chunk.Blocks[x][y][z].Type = BlockStone
					} else if y <= 33 {
						chunk.Blocks[x][y][z].Type = BlockSand
					} else {
						chunk.Blocks[x][y][z].Type = BlockGrass
					}
				} else if y > heightInt-4 {
					if y > 80 {
						chunk.Blocks[x][y][z].Type = BlockStone
					} else if heightInt <= 33 {
						chunk.Blocks[x][y][z].Type = BlockSand
					} else {
						chunk.Blocks[x][y][z].Type = BlockDirt
					}
				} else {
					chunk.Blocks[x][y][z].Type = BlockStone
				}
			}
		}
	}
	return chunk
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

	// Update neighboring chunks if block is on edge
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
			if w.chunks[key].Mesh != nil {
				gl.DeleteVertexArrays(1, &w.chunks[key].Mesh.VAO)
				gl.DeleteBuffers(1, &w.chunks[key].Mesh.VBO)
			}
			toDelete = append(toDelete, key)
		}
	}

	for _, key := range toDelete {
		delete(w.chunks, key)
	}
}
