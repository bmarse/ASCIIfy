// Package asciify holds all the logic for converting images into fun ascii art
package asciify

import (
	"image"
	"math"
	"sync"

	_ "image/jpeg"
	_ "image/png"

	"github.com/nfnt/resize"
)

const SHADER = " .:coPO?@#"

func CreateScreenBuffer(w int, h int) []byte {
	row := w + 1
	buf := make([]byte, h*row)
	for i := range buf {
		buf[i] = ' '
	}

	// line endings
	for i := range h {
		buf[(i+1)*row-1] = '\n'
	}

	return buf
}

func ResizeImage(m image.Image, w uint, h uint) image.Image {
	return resize.Thumbnail(w, h, m, resize.Lanczos3)
}

func ImageToASCIIToBuf(img image.Image, w int, h int, buf []byte) {
	// Render one output character per output column to avoid overwriting newline bytes.
	// Each two output columns sample the same source pixel; compute source width as ceil(w/2).
	sourceWidth := (w + 1) / 2
	img = ResizeImage(img, uint(sourceWidth), uint(h))

	bounds := img.Bounds()
	minX, minY := bounds.Min.X, bounds.Min.Y
	maxX, maxY := bounds.Max.X, bounds.Max.Y
	rowStride := w + 1

	var wg sync.WaitGroup
	wg.Add(h)

	for y := range h {
		// capture loop var
		go func(y int) {
			defer wg.Done()
			lineStart := y * rowStride
			for x := range w {
				ix := minX + x/2
				iy := minY + y
				if ix >= maxX || iy >= maxY {
					buf[lineStart+x] = SHADER[0]
					continue
				}

				r, g, b, _ := img.At(ix, iy).RGBA()
				// Calc lum https://en.wikipedia.org/wiki/Relative_luminance
				rf := 0.2126 * float64(r)
				gf := 0.7152 * float64(g)
				bf := 0.0722 * float64(b)
				lum := int(math.Floor(((rf + gf + bf) / 0xffff) * float64(len(SHADER)-1)))
				if lum < 0 {
					lum = 0
				} else if lum >= len(SHADER) {
					lum = len(SHADER) - 1
				}
				buf[lineStart+x] = SHADER[lum]
			}
		}(y)
	}

	wg.Wait()
}
