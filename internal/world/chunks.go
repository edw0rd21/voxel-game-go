package world

import (
	"github.com/go-gl/gl/v4.1-core/gl"
)

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

func (c *Chunk) generateMesh(w *World) {
	vertices := make([]float32, 0, 4096)

	// Cache neighbors to avoid map lookups in the inner loop
	nLeft := w.chunks[[2]int{c.X - 1, c.Z}]
	nRight := w.chunks[[2]int{c.X + 1, c.Z}]
	nBack := w.chunks[[2]int{c.X, c.Z - 1}]
	nFront := w.chunks[[2]int{c.X, c.Z + 1}]

	// Helper closure to check transparency
	isTransparent := func(x, y, z int) bool {
		if y < 0 || y >= ChunkHeight {
			return true
		}
		if x >= 0 && x < ChunkSize && z >= 0 && z < ChunkSize {
			return c.Blocks[x][y][z].Type == BlockAir
		}
		// Neighbor checks
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

				wx := float32(c.X*ChunkSize + x)
				wy := float32(y)
				wz := float32(c.Z*ChunkSize + z)

				// Face checks
				if isTransparent(x, y, z+1) {
					addFace(&vertices, wx, wy, wz, 0, blockType) // Front
				}
				if isTransparent(x, y, z-1) {
					addFace(&vertices, wx, wy, wz, 1, blockType) // Back
				}
				if isTransparent(x+1, y, z) {
					addFace(&vertices, wx, wy, wz, 2, blockType) // Right
				}
				if isTransparent(x-1, y, z) {
					addFace(&vertices, wx, wy, wz, 3, blockType) // Left
				}
				if isTransparent(x, y+1, z) {
					addFace(&vertices, wx, wy, wz, 4, blockType) // Top
				}
				if isTransparent(x, y-1, z) {
					addFace(&vertices, wx, wy, wz, 5, blockType) // Bottom
				}
			}
		}
	}

	if len(vertices) == 0 {
		return
	}

	if c.Mesh == nil {
		c.Mesh = &ChunkMesh{}
		gl.GenVertexArrays(1, &c.Mesh.VAO)
		gl.GenBuffers(1, &c.Mesh.VBO)
	}

	gl.BindVertexArray(c.Mesh.VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, c.Mesh.VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	// Stride is 8 floats: X,Y,Z (3) + U,V (2) + Nx,Ny,Nz (3)
	stride := int32(8 * 4)

	// Position (3 floats)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, stride, gl.PtrOffset(0))

	// TexCoord (2 floats) -- REPLACES COLOR
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, stride, gl.PtrOffset(3*4))

	// Normal (3 floats)
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, stride, gl.PtrOffset(5*4))

	gl.BindVertexArray(0)
	c.Mesh.VertexCount = len(vertices) / 8 // 8 floats per vertex
}

func addFace(verts *[]float32, x, y, z float32, face int, bType BlockType) {
	// Get UV coordinates for this specific face
	u, v := GetBlockUVs(bType, face)

	// Determine Normals based on face
	var nx, ny, nz float32
	switch face {
	case 0:
		nz = 1 // Front
	case 1:
		nz = -1 // Back
	case 2:
		nx = 1 // Right
	case 3:
		nx = -1 // Left
	case 4:
		ny = 1 // Top
	case 5:
		ny = -1 // Bottom
	}

	// Append Quad (2 Triangles)
	// Format: X, Y, Z, U, V, Nx, Ny, Nz

	// Helper to reduce typing
	appendVert := func(vx, vy, vz, vu, vv float32) {
		*verts = append(*verts, vx, vy, vz, vu, vv, nx, ny, nz)
	}

	uSize := float32(TileSize) / float32(TextureWidth)
	vSize := float32(TileSize) / float32(TextureHeight) // Width of one tile in UV space

	if face == 0 { // Front (+Z)
		appendVert(x, y, z+1, u, v+vSize)         // Bottom Left
		appendVert(x+1, y, z+1, u+uSize, v+vSize) // Bottom Right
		appendVert(x+1, y+1, z+1, u+uSize, v)     // Top Right
		appendVert(x, y, z+1, u, v+vSize)         // Bottom Left
		appendVert(x+1, y+1, z+1, u+uSize, v)     // Top Right
		appendVert(x, y+1, z+1, u, v)             // Top Left
	} else if face == 1 { // Back (-Z)
		// Note: UVs often need flipping depending on your specific atlas
		// We use standard mapping here
		appendVert(x+1, y, z, u, v+vSize)
		appendVert(x, y, z, u+uSize, v+vSize)
		appendVert(x, y+1, z, u+uSize, v) // Top Right
		appendVert(x+1, y, z, u, v+vSize)
		appendVert(x, y+1, z, u+uSize, v)
		appendVert(x+1, y+1, z, u, v)
	} else if face == 2 { // Right (+X)
		appendVert(x+1, y, z+1, u, v+vSize)
		appendVert(x+1, y, z, u+uSize, v+vSize)
		appendVert(x+1, y+1, z, u+uSize, v) // Top Right
		appendVert(x+1, y, z+1, u, v+vSize)
		appendVert(x+1, y+1, z, u+uSize, v)
		appendVert(x+1, y+1, z+1, u+uSize, v) // Top Right
	} else if face == 3 { // Left (-X)
		appendVert(x, y, z, u, v+vSize)
		appendVert(x, y, z+1, u+uSize, v+vSize)
		appendVert(x, y+1, z+1, u+uSize, v) // Top Right
		appendVert(x, y, z, u, v+vSize)
		appendVert(x, y+1, z+1, u+uSize, v)
		appendVert(x, y+1, z, u, v)
	} else if face == 4 { // Top (+Y)
		appendVert(x, y+1, z+1, u, v+vSize)
		appendVert(x+1, y+1, z+1, u+uSize, v+vSize)
		appendVert(x+1, y+1, z, u+uSize, v)
		appendVert(x, y+1, z+1, u, v+vSize)
		appendVert(x+1, y+1, z, u+uSize, v)
		appendVert(x, y+1, z, u, v)
	} else if face == 5 { // Bottom (-Y)
		appendVert(x, y, z, u, v+vSize)
		appendVert(x+1, y, z, u+uSize, v+vSize)
		appendVert(x+1, y, z+1, u+uSize, v)
		appendVert(x, y, z, u, v+vSize)
		appendVert(x+1, y, z+1, u+uSize, v)
		appendVert(x, y, z+1, u, v)
	}
}
