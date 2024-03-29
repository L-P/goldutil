package sprite

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"unsafe" // informational sizeof
)

type Type int32

const (
	ParallelUpright Type = iota
	FacingUpright
	Parallel
	Oriented
	ParallelOriented
)

func (typ Type) String() string {
	switch typ {
	case ParallelUpright:
		return "ParallelUpright"
	case FacingUpright:
		return "FacingUpright"
	case Parallel:
		return "Parallel"
	case Oriented:
		return "Oriented"
	case ParallelOriented:
		return "ParallelOriented"
	default:
		return fmt.Sprintf("invalid (%d)", typ)
	}
}

type TextureFormat int32

const (
	Normal TextureFormat = iota
	Additive
	IndexAlpha
	AlphaTest
)

func (tf TextureFormat) String() string {
	switch tf {
	case Normal:
		return "Normal"
	case Additive:
		return "Additive"
	case IndexAlpha:
		return "IndexAlpha"
	case AlphaTest:
		return "AlphaTest"
	default:
		return fmt.Sprintf("invalid (%d)", tf)
	}
}

type SyncType int32

const (
	Sync SyncType = iota
	Random
)

func (st SyncType) String() string {
	switch st {
	case Sync:
		return "Sync"
	case Random:
		return "Random"
	default:
		return fmt.Sprintf("invalid (%d)", st)
	}
}

const expectedPaletteSize = 256

// Binary-accurate.
type Header struct {
	MagicString    [4]byte // "IDSP"
	Version        int32   // 1 for Quake, 2 for Valve
	Type           Type
	TextureFormat  TextureFormat
	BoundingRadius float32 // radius of the bounding sphere (sqrt(w/2² + h/2²))
	Width, Height  int32
	NumFrames      int32

	// The quake engine scales by minus this number before rotating a sprite.
	// Probably unused.
	BeamLength int32

	// Informs the quake engine to desync client-side animations (makes them
	// start with a random delay).
	SyncType SyncType

	// The palette is a Valve addition in sprite format version 2.
	PaletteSize int16   // always 256
	Palette     Palette // always 3 bytes * PaletteSize, keep it fixed to simplify parsing
}

func (sh *Header) Read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, sh); err != nil {
		return err
	}

	var IDSP = [4]byte{'I', 'D', 'S', 'P'}
	if sh.MagicString != IDSP {
		return errors.New("cannot find magic string, probably not a sprite file")
	}

	if sh.Version != 2 {
		return fmt.Errorf("unhandled sprite version %d, expected version 2", sh.Version)
	}

	if sh.PaletteSize != expectedPaletteSize {
		return fmt.Errorf(
			"unhandled palette size: %d, expected %d",
			sh.PaletteSize, expectedPaletteSize,
		)
	}

	return nil
}

func (sh Header) String() string {
	var w strings.Builder

	w.WriteString("Header:\n")
	fmt.Fprintf(&w, "  MagicString: %s\n", sh.MagicString)
	fmt.Fprintf(&w, "  Version: %d\n", sh.Version)
	fmt.Fprintf(&w, "  Type: %s\n", sh.Type.String())
	fmt.Fprintf(&w, "  TextureFormat: %s\n", sh.TextureFormat.String())
	fmt.Fprintf(&w, "  BoundingRadius: %f\n", sh.BoundingRadius)
	fmt.Fprintf(&w, "  Width: %d\n", sh.Width)
	fmt.Fprintf(&w, "  Height: %d\n", sh.Height)
	fmt.Fprintf(&w, "  NumFrames: %d\n", sh.NumFrames)
	fmt.Fprintf(&w, "  BeamLength: %d\n", sh.BeamLength)
	fmt.Fprintf(&w, "  SyncType: %s\n", sh.SyncType.String())
	fmt.Fprintf(&w, "  PaletteSize: %d\n", sh.PaletteSize)
	fmt.Fprintf(&w, "  Palette: %d bytes\n", len(sh.Palette)*int(unsafe.Sizeof(RGB{})))

	return w.String()
}

func (sh Header) Write(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, sh)
}
