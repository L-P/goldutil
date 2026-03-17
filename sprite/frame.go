package sprite

import (
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"strings"
)

type FrameType int32

const (
	Single FrameType = iota
	Group            // not implemented in GoldSrc.
)

func (ft FrameType) String() string {
	switch ft {
	case Single:
		return "Single"
	case Group:
		return "Group"
	default:
		return fmt.Sprintf("invalid (%d)", ft)
	}
}

type Frame struct {
	// split out to simplify calling binary.Read, sorry.
	FrameMeta

	Data []byte // Width*Height, raw 8bpp indexed
}

type FrameMeta struct {
	Type             FrameType
	OriginX, OriginY int32
	Width, Height    int32
}

func NewFrame(
	width, height int32,
	originX, originY int32,
	data []byte,
) Frame {
	return Frame{
		FrameMeta: FrameMeta{
			Type:    Single,
			Width:   width,
			Height:  height,
			OriginX: originX,
			OriginY: originY,
		},
		Data: data,
	}
}

func (f *Frame) Read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &f.FrameMeta); err != nil {
		return fmt.Errorf("unable to parse frame meta-data: %w", err)
	}

	if f.Type != Single {
		return fmt.Errorf("unhandled frame type: %s", f.Type.String())
	}

	if f.Width <= 0 || f.Height <= 0 {
		return fmt.Errorf("invalid dimensions: %d×%d", f.Width, f.Height)
	}

	f.Data = make([]byte, f.Width*f.Height)
	if err := binary.Read(r, binary.LittleEndian, &f.Data); err != nil {
		return fmt.Errorf("unable to frame data: %w", err)
	}

	return nil
}

func (f *Frame) String() string {
	var w strings.Builder

	fmt.Fprintf(&w, "  Type: %s\n", f.Type.String())
	fmt.Fprintf(&w, "  OriginX: %d\n", f.OriginX)
	fmt.Fprintf(&w, "  OriginY: %d\n", f.OriginY)
	fmt.Fprintf(&w, "  Width: %d\n", f.Width)
	fmt.Fprintf(&w, "  Height: %d\n", f.Height)
	fmt.Fprintf(&w, "  Data: %d bytes\n", len(f.Data))

	return w.String()
}

func (f *Frame) Rect() image.Rectangle {
	return image.Rect(0, 0, int(f.Width), int(f.Height))
}

func (f *Frame) Write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, f.FrameMeta); err != nil {
		return fmt.Errorf("unable to write frame header: %w", err)
	}

	if _, err := w.Write(f.Data); err != nil {
		return fmt.Errorf("unable to write frame data: %w", err)
	}

	return nil
}
