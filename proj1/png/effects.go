package png

import "image/color"

var kernels = map[string][9]float64{
	"S": {0, -1, 0, -1, 5, -1, 0, -1, 0},
	"E": {-1, -1, -1, -1, 8, -1, -1, -1, -1},
	"B": {1.0 / 9, 1.0 / 9, 1.0 / 9, 1.0 / 9, 1.0 / 9, 1.0 / 9, 1.0 / 9, 1.0 / 9, 1.0 / 9},
}

// ApplyRows reads the old pixels and writes only rows [start, end).
// Workers share the input image, but write to different output rows.
func (img *Image) ApplyRows(effect string, start, end int) {
	kernel, ok := kernels[effect]
	if effect != "G" && !ok {
		panic("unknown effect: " + effect)
	}
	for y := start; y < end; y++ {
		for x := img.Bounds.Min.X; x < img.Bounds.Max.X; x++ {
			pixel := img.in.RGBA64At(x, y)
			if effect == "G" {
				gray := clamp((float64(pixel.R) + float64(pixel.G) + float64(pixel.B)) / 3)
				img.out.SetRGBA64(x, y, color.RGBA64{gray, gray, gray, pixel.A})
				continue
			}
			var r, g, b float64
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					// RGBA64At returns zero outside the full image. It still
					// reads real pixels across a worker's slice boundary.
					neighbor := img.in.RGBA64At(x+dx, y+dy)
					weight := kernel[(dy+1)*3+dx+1]
					r += float64(neighbor.R) * weight
					g += float64(neighbor.G) * weight
					b += float64(neighbor.B) * weight
				}
			}
			img.out.SetRGBA64(x, y, color.RGBA64{clamp(r), clamp(g), clamp(b), pixel.A})
		}
	}
}

// Swap makes the finished output the input for the next effect.
// Call this only after all workers have finished writing.
func (img *Image) Swap() {
	img.in, img.out = img.out, img.in
}

// Grayscale keeps the provided sample program working.
func (img *Image) Grayscale() {
	img.ApplyRows("G", img.Bounds.Min.Y, img.Bounds.Max.Y)
	img.Swap()
}
