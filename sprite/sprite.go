package sprite

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

const expectedPaletteSize = 256

// .spr file, originally from Quake and modified by Valve.
// See sprgen.c and the Quake engine.
type Sprite struct {
	Header

	Palette [3 * expectedPaletteSize]byte // always 3 * PaletteSize
	Frames  []Frame
}

func (spr Sprite) String() string {
	var w strings.Builder

	w.WriteString(spr.Header.String())
	fmt.Fprintf(&w, "Palette: %d bytes\n", len(spr.Palette))

	for i, v := range spr.Frames {
		fmt.Fprintf(&w, "Frame %d:\n", i)
		w.WriteString(v.String())
	}

	return w.String()
}

func (spr *Sprite) Read(r io.Reader) error {
	if err := spr.Header.Read(r); err != nil {
		return fmt.Errorf("could not parse header: %w", err)
	}

	if spr.PaletteSize != expectedPaletteSize {
		return fmt.Errorf(
			"unhandled palette size: %d, expected %d",
			spr.PaletteSize, expectedPaletteSize,
		)
	}

	if err := binary.Read(r, binary.LittleEndian, &spr.Palette); err != nil {
		return fmt.Errorf("could not parse palette: %w", err)
	}

	spr.Frames = make([]Frame, 0, spr.NumFrames)
	for i := int32(0); i < spr.NumFrames; i += 1 {
		var frame Frame
		if err := frame.Read(r); err != nil {
			return fmt.Errorf("unable to read frame %d: %w", i, err)
		}

		spr.Frames = append(spr.Frames, frame)
	}

	return nil
}

func NewFromFile(path string) (Sprite, error) {
	f, err := os.Open(path)
	if err != nil {
		return Sprite{}, fmt.Errorf("unable to open file: %w", err)
	}
	defer f.Close()

	var sprite Sprite
	if err := sprite.Read(f); err != nil {
		return Sprite{}, err
	}

	return sprite, nil
}
