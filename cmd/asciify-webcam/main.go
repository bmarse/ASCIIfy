package main

import (
	"context"
	"fmt"
	"image"
	"os"
	"time"

	"github.com/bmarse/ascii-render/pkg/asciify"
	"github.com/gosuri/uilive"
	"github.com/svanichkin/gocam"
	"golang.org/x/term"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	frames, err := gocam.StartStream(ctx)
	if err != nil {
		panic(err)
	}

	// Get terminal size for ASCII output
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	screen := uilive.New()
	screen.RefreshInterval = time.Millisecond * 50
	screen.Start()
	defer screen.Stop()

	c := make(chan []byte, 5)

	// Producer: read frames, convert to RGBA, render ASCII into buffers and send
	go func() {
		for frame := range frames {
			if frame.Width == 0 || frame.Height == 0 {
				continue
			}
			total := frame.Width * frame.Height
			var img *image.RGBA
			expectedLen := total * 3
			if total > 0 && len(frame.Data) >= expectedLen {
				img = image.NewRGBA(image.Rect(0, 0, frame.Width, frame.Height))
				offset := 0
				for p := 0; p < total && offset+2 < len(frame.Data); p++ {
					y := float32(frame.Data[offset+0])
					cb := float32(frame.Data[offset+1]) - 128.0
					cr := float32(frame.Data[offset+2]) - 128.0

					fr := y + 1.40200*cr
					fg := y - 0.344136*cb - 0.714136*cr
					fb := y + 1.77200*cb

					clamp := func(v float32) uint8 {
						if v < 0 {
							return 0
						}
						if v > 255 {
							return 255
						}
						return uint8(v + 0.5)
					}

					idx := p * 4
					img.Pix[idx+0] = clamp(fr)
					img.Pix[idx+1] = clamp(fg)
					img.Pix[idx+2] = clamp(fb)
					img.Pix[idx+3] = 255

					offset += 3
				}
			}

			if img == nil {
				continue
			}

			// Create a buffer sized to the terminal and render into it.
			// Make a fresh buffer per frame to avoid races.
			buf := asciify.CreateScreenBuffer(w, h)
			asciify.ImageToASCIIToBuf(img, w, h, buf)
			c <- buf
		}
		close(c)
	}()

	// Consumer: write buffers to the screen at a fixed tick rate (approx 10 FPS)
	ticker := time.NewTicker(time.Millisecond * 10)
	defer ticker.Stop()

	for {
		b, ok := <-c
		if !ok {
			break
		}
		<-ticker.C
		if _, err := screen.Write(b); err != nil {
			fmt.Fprintln(os.Stderr, "screen write error:", err)
			break
		}
		if err := screen.Flush(); err != nil {
			fmt.Fprintln(os.Stderr, "screen flush error:", err)
			break
		}
	}
}
