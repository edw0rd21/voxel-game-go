package world

type BlockType uint8

// Block Types
const (
	BlockAir   = 0
	BlockDirt  = 1
	BlockGrass = 2
	BlockStone = 3
	BlockSnow  = 4
)

// Texture Atlas Constants
const (
	TextureWidth  = 1152.0
	TextureHeight = 1280.0
	TileSize      = 128.0
)

// Texture Indice
var (
	TexDirt      = [2]float32{6, 0}
	TexGrassTop  = [2]float32{8, 0}
	TexGrassSide = [2]float32{7, 4}
	TexStone     = [2]float32{3, 4}
	TexSnow      = [2]float32{3, 5}
)

// Texture Coordinates helper
func GetBlockUVs(blockType BlockType, faceDirection int) (float32, float32) {
	var tileCoords [2]float32

	switch blockType {
	case BlockDirt:
		tileCoords = TexDirt
	case BlockStone:
		tileCoords = TexStone
	case BlockSnow:
		tileCoords = TexSnow
	case BlockGrass:
		if faceDirection == 4 { // Top
			tileCoords = TexGrassTop
		} else if faceDirection == 5 { // Bottom
			tileCoords = TexDirt
		} else {
			tileCoords = TexGrassSide
		}
	default:
		tileCoords = [2]float32{0, 0}
	}

	pixelX := (tileCoords[0] * TileSize)
	pixelY := (tileCoords[1] * TileSize)

	u := pixelX / TextureWidth
	v := pixelY / TextureHeight
	return u, v
}
