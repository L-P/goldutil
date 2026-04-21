package sprite

import (
	"encoding/binary"
	"errors"
	"fmt"
	"goldutil/palette"
	"io"
	"strings"
	"unsafe" // informational sizeof
)

// Binary-accurate.
type Type int32

const TypeInvalid Type = -1
const (
	TypeParallelUpright Type = iota
	TypeFacingUpright
	TypeParallel
	TypeOriented
	TypeParallelOriented
)

// Parse a type from our own human-readable representation.
func ParseType(str string) (Type, error) {
	switch str {
	case "parallel-upright":
		return TypeParallelUpright, nil
	case "facing-upright":
		return TypeFacingUpright, nil
	case "parallel":
		return TypeParallel, nil
	case "oriented":
		return TypeOriented, nil
	case "parallel-oriented":
		return TypeParallelOriented, nil
	}

	return TypeInvalid, fmt.Errorf("unrecognize sprite type: %s", str)
}

func (typ Type) String() string {
	switch typ {
	case TypeParallelUpright:
		return "ParallelUpright"
	case TypeFacingUpright:
		return "FacingUpright"
	case TypeParallel:
		return "Parallel"
	case TypeOriented:
		return "Oriented"
	case TypeParallelOriented:
		return "ParallelOriented"
	case TypeInvalid:
		fallthrough
	default:
		return fmt.Sprintf("invalid (%d)", typ)
	}
}

// Binary-accurate.
type TextureFormat int32

// Parse a texture format from our own human-readable representation.
func ParseTextureFormat(str string) (TextureFormat, error) {
	switch str {
	case "normal":
		return TextureFormatNormal, nil
	case "additive":
		return TextureFormatAdditive, nil
	case "index-alpha":
		return TextureFormatIndexAlpha, nil
	case "alpha-test":
		return TextureFormatAlphaTest, nil
	}

	return TextureFormatInvalid, fmt.Errorf("unrecognize texture format: %s", str)
}

const TextureFormatInvalid TextureFormat = -1
const (
	TextureFormatNormal TextureFormat = iota
	TextureFormatAdditive
	TextureFormatIndexAlpha
	TextureFormatAlphaTest
)

func (tf TextureFormat) String() string {
	switch tf {
	case TextureFormatNormal:
		return "Normal"
	case TextureFormatAdditive:
		return "Additive"
	case TextureFormatIndexAlpha:
		return "IndexAlpha"
	case TextureFormatAlphaTest:
		return "AlphaTest"
	case TextureFormatInvalid:
		fallthrough
	default:
		return fmt.Sprintf("invalid (%d)", tf)
	}
}

type SyncType int32

const (
	SyncTypeSynced SyncType = iota
	SyncTypeRandom
)

func (st SyncType) String() string {
	switch st {
	case SyncTypeSynced:
		return "Synced"
	case SyncTypeRandom:
		return "Random"
	default:
		return fmt.Sprintf("invalid (%d)", st)
	}
}

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
	PaletteSize int16           // always 256
	Palette     palette.Palette // always 3 bytes * PaletteSize, keep it fixed to simplify parsing
}

func (sh *Header) read(r io.Reader) error {
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

	if sh.PaletteSize != palette.Size {
		return fmt.Errorf(
			"unhandled palette size: %d, expected %d",
			sh.PaletteSize, palette.Size,
		)
	}

	return nil
}

func (sh *Header) String() string {
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
	fmt.Fprintf(&w, "  Palette: %d bytes\n", len(sh.Palette)*int(unsafe.Sizeof(palette.RGB{})))

	return w.String()
}

func (sh *Header) Write(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, sh)
}
