package ui

import (
	"fmt"
	"image"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type GlyphInfo struct {
	UVMin   mgl32.Vec2
	UVMax   mgl32.Vec2
	Size    mgl32.Vec2
	Bearing mgl32.Vec2
	Advance float32
}

type Font struct {
	TextureID  uint32
	Glyphs     map[rune]GlyphInfo
	LineHeight float32
}

func LoadFont(filePath string, fontSize float64) (*Font, error) {
	// Read font file
	fontBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read font file: %w", err)
	}

	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("could not parse font: %w", err)
	}

	// Setup Atlas Image (512x512 for basic ASCII)
	const atlasSize = 512
	const padding = 2
	atlasImg := image.NewRGBA(image.Rect(0, 0, atlasSize, atlasSize))

	// Context for drawing text
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(fontSize)
	c.SetClip(atlasImg.Bounds())
	c.SetDst(atlasImg)
	c.SetSrc(image.White) // Draw white text
	c.SetHinting(font.HintingFull)

	// Create a Face to access low-level metrics
	opts := truetype.Options{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	}
	face := truetype.NewFace(f, &opts)

	// Calculate global metrics
	metrics := face.Metrics()
	// Convert fixed.Int26_6 to float32
	ascent := float32(metrics.Ascent.Ceil())
	descent := float32(metrics.Descent.Ceil())
	lineHeight := ascent + descent

	// Render Glyphs
	glyphs := make(map[rune]GlyphInfo)

	currentX := padding
	currentY := padding
	maxRowHeight := 0

	// Render ASCII range 32 (space) to 126 (~)
	for ch := rune(32); ch <= 126; ch++ {
		b, advance, ok := face.GlyphBounds(ch)
		if !ok {
			continue // Glyph missing
		}

		gw := (b.Max.X - b.Min.X).Ceil()
		gh := (b.Max.Y - b.Min.Y).Ceil()

		if currentX+gw+padding >= atlasSize {
			currentX = padding
			currentY += maxRowHeight + padding
			maxRowHeight = 0 // Reset row height
		}

		if currentY+gh+padding >= atlasSize {
			return nil, fmt.Errorf("font atlas full (increase atlasSize or reduce fontSize)")
		}
		dotX := currentX - b.Min.X.Floor()
		dotY := currentY - b.Min.Y.Floor()

		pt := fixed.P(dotX, dotY)
		c.DrawString(string(ch), pt)

		// Normalize pixel coordinates to 0.0-1.0 range
		uMin := float32(currentX) / float32(atlasSize)
		vMin := float32(currentY) / float32(atlasSize)
		uMax := float32(currentX+gw) / float32(atlasSize)
		vMax := float32(currentY+gh) / float32(atlasSize)

		bearingX := float32(b.Min.X) / 64.0
		bearingY := float32(b.Max.Y) / 64.0

		glyphs[ch] = GlyphInfo{
			UVMin:   mgl32.Vec2{uMin, vMin},
			UVMax:   mgl32.Vec2{uMax, vMax},
			Size:    mgl32.Vec2{float32(gw), float32(gh)},
			Bearing: mgl32.Vec2{bearingX, bearingY},
			Advance: float32(advance) / 64.0,
		}

		// Advance packing cursor
		currentX += gw + padding
		if gh > maxRowHeight {
			maxRowHeight = gh
		}
	}

	//Upload Texture
	var texID uint32
	gl.GenTextures(1, &texID)
	gl.BindTexture(gl.TEXTURE_2D, texID)

	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA, // Keeping RGBA for shader compatibility
		int32(atlasSize),
		int32(atlasSize),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(atlasImg.Pix),
	)

	// Linear filtering for smooth text
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.BindTexture(gl.TEXTURE_2D, 0)

	return &Font{
		TextureID:  texID,
		Glyphs:     glyphs,
		LineHeight: lineHeight,
	}, nil

}
