package asciify

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateScreenBuffer(t *testing.T) {
	w, h := 4, 2
	buf := CreateScreenBuffer(w, h)
	assert.Equal(t, []byte{32, 32, 32, 32, 10, 32, 32, 32, 32, 10}, buf)
}

func TestImageToASCIIToBuf_SimpleMapping(t *testing.T) {
	// Create a 2x2 source image. We will render to a buffer with width=4, height=2.
	// ImageToASCIIToBuf duplicates horizontal pixels (x += 2 -> write two chars per source pixel).
	w, h := 4, 2
	buf := CreateScreenBuffer(w, h)

	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	// Layout:
	// (0,0) white, (1,0) black
	// (0,1) white, (1,1) white
	img.Set(0, 0, color.NRGBA{255, 255, 255, 255})
	img.Set(1, 0, color.NRGBA{0, 0, 0, 255})
	img.Set(0, 1, color.NRGBA{255, 255, 255, 255})
	img.Set(1, 1, color.NRGBA{255, 255, 255, 255})

	ImageToASCIIToBuf(img, w, h, buf)

	got := string(buf)
	// For white pixels we expect SHADER[len-1] duplicated ('#'), for black SHADER[0] which is space.
	// Row 0: white -> "##", black -> "  "  => "##  \n"
	// Row 1: white -> "##", white -> "##" => "####\n"
	expected := "##  \n####\n"

	assert.Equal(t, expected, got, "unexpected ascii output")
}
