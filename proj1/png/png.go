// Package png loads PNG files and applies image effects.
package png

import (
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
)

type Image struct {
	in     *image.RGBA64 // Current pixels. Workers only read this buffer.
	out    *image.RGBA64 // Space for the next result.
	Bounds image.Rectangle
}

func Load(path string) (*Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	original, err := png.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := original.Bounds()
	input := image.NewRGBA64(bounds)
	// Convert the decoded image to our RGBA64 buffer once.
	draw.Draw(input, bounds, original, bounds.Min, draw.Src)
	return &Image{in: input, out: image.NewRGBA64(bounds), Bounds: bounds}, nil
}

// After a swap, in holds the latest result. With no effects it is the original.
func (img *Image) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	// Faster lossless compression saves CPU time without changing pixels.
	encoder := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := encoder.Encode(file, img.in); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func clamp(value float64) uint16 {
	return uint16(math.Min(65535, math.Max(0, value)))
}
