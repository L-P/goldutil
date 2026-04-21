// Package sprite implements GoldSrc SPR files parsing.
package sprite

import (
	"fmt"
	"github.com/L-P/goldutil/palette"
	"io"
	"math"
	"os"
	"strings"
)

// .spr file, originally from Quake and modified by Valve.
// See sprgen.c and the Quake engine.
type Sprite struct {
	Header

	Frames []Frame
}

func New(
	width, height int, typ Type, format TextureFormat,
	pal palette.Palette,
) (Sprite, error) {
	var spr Sprite
	spr.Header = Header{
		MagicString:    [4]byte{'I', 'D', 'S', 'P'},
		Version:        2,
		Type:           typ,
		TextureFormat:  format,
		BoundingRadius: boundingRadius(width, height),
		Width:          int32(width),
		Height:         int32(height),
		PaletteSize:    palette.Size,
		Palette:        pal,
	}

	return spr, nil
}

func boundingRadius(iWidth, iHeight int) float32 {
	return float32(math.Sqrt(
		math.Pow(float64(iWidth)/2, 2) + math.Pow(float64(iHeight)/2, 2),
	))
}

func (spr *Sprite) String() string {
	var w strings.Builder

	w.WriteString(spr.Header.String())

	for i, v := range spr.Frames {
		fmt.Fprintf(&w, "Frame %d:\n", i)
		w.WriteString(v.String())
	}

	return w.String()
}

func (spr *Sprite) read(r io.Reader) error {
	if err := spr.Header.read(r); err != nil {
		return fmt.Errorf("could not parse header: %w", err)
	}

	spr.Frames = make([]Frame, 0, spr.NumFrames)
	for i := range spr.NumFrames {
		var frame Frame
		if err := frame.Read(r); err != nil {
			return fmt.Errorf("unable to read frame %d: %w", i, err)
		}

		spr.Frames = append(spr.Frames, frame)
	}

	return nil
}

func NewFromReader(r io.Reader) (Sprite, error) {
	var sprite Sprite
	if err := sprite.read(r); err != nil {
		return Sprite{}, err
	}

	return sprite, nil
}

func NewFromFile(path string) (Sprite, error) {
	f, err := os.Open(path)
	if err != nil {
		return Sprite{}, fmt.Errorf("unable to open file: %w", err)
	}
	defer f.Close() //nolint:errcheck // readonly

	return NewFromReader(f)
}

func (spr *Sprite) AddFrame(frame Frame) {
	spr.Frames = append(spr.Frames, frame)
	spr.NumFrames++
}

func (spr *Sprite) Write(w io.Writer) error {
	if err := spr.Header.Write(w); err != nil {
		return err
	}

	for i, frame := range spr.Frames {
		if err := frame.Write(w); err != nil {
			return fmt.Errorf("could not write frame #%d: %w", i, err)
		}
	}

	return nil
}
