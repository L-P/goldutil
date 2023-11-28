package sprite

import (
	"encoding/binary"
	"fmt"
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

func (fd *Frame) Read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &fd.FrameMeta); err != nil {
		return fmt.Errorf("unable to parse frame meta-data: %w", err)
	}

	if fd.Type != Single {
		return fmt.Errorf("unhandled frame type: %s", fd.Type.String())
	}

	if fd.Width <= 0 || fd.Height <= 0 {
		return fmt.Errorf("invalid dimensions: %d×%d", fd.Width, fd.Height)
	}

	fd.Data = make([]byte, fd.Width*fd.Height)
	if err := binary.Read(r, binary.LittleEndian, &fd.Data); err != nil {
		return fmt.Errorf("unable to frame data: %w", err)
	}

	return nil
}

func (f Frame) String() string {
	var w strings.Builder

	fmt.Fprintf(&w, "  Type: %s\n", f.Type.String())
	fmt.Fprintf(&w, "  OriginX: %d\n", f.OriginX)
	fmt.Fprintf(&w, "  OriginY: %d\n", f.OriginY)
	fmt.Fprintf(&w, "  Width: %d\n", f.Width)
	fmt.Fprintf(&w, "  Height: %d\n", f.Height)
	fmt.Fprintf(&w, "  Data: %d bytes\n", len(f.Data))

	return w.String()
}
