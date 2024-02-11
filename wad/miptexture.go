package wad

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// Number of mimmaps per texture, base texture is mipmap 0.
const NumMIPMaps = 4

type RGB struct {
	R, G, B uint8
}

type Palette [256]RGB

// Binary-accurate.
type MIPTextureHeader struct {
	// Seems to be a lowercase version of the dir entry name. Unused?
	Name TextureName

	// multiples of 16 (guarantees 4 half-size mipmaps)
	Width  int32
	Height int32

	MIPOffsets [NumMIPMaps]int32
}

type MIPTexture struct {
	MIPTextureHeader

	// total len = w*h + (w*h)/2 + (w*h)/4 + (w*h)/8 = (15*(w*h))/8
	// Paletted data, 1bpp.
	MIPData [NumMIPMaps][]byte
	_       [2]byte
	Palette Palette
	_       [2]byte
}

func (mip MIPTexture) String() string {
	var w strings.Builder
	fmt.Fprintf(&w, "  Name: %s\n", mip.Name.String())
	fmt.Fprintf(&w, "  Width: %d\n", mip.Width)
	fmt.Fprintf(&w, "  Height: %d\n", mip.Height)

	for i := range mip.MIPOffsets {
		fmt.Fprintf(&w, "  MIPMap #%d Offset: 0x%x\n", i, mip.MIPOffsets[i])
		fmt.Fprintf(&w, "  MIPMap #%d Size: %d\n", i, len(mip.MIPData[i]))
	}

	return w.String()
}

func (mip *MIPTexture) Read(r io.ReadSeeker, offset, size int32) error {
	if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek to offset %x of MIPTexture header", offset)
	}

	if err := binary.Read(r, binary.LittleEndian, &mip.MIPTextureHeader); err != nil {
		return fmt.Errorf("unable to read MIPTextureHeader: %w", err)
	}

	for i := range mip.MIPData {
		// Each MIPMap is half the size of the previous one.
		size := (mip.Width * mip.Height) / ((2 << i) / 2)
		mip.MIPData[i] = make([]byte, size)

		mipmapOffset := int64(offset + mip.MIPOffsets[i])
		if _, err := r.Seek(mipmapOffset, io.SeekStart); err != nil {
			return fmt.Errorf("unable to seek to MIPMap data #%d: %w", i, err)
		}

		if err := binary.Read(r, binary.LittleEndian, &mip.MIPData[i]); err != nil {
			return fmt.Errorf("unable to read MIPMap #%d: %w", i, err)
		}
	}

	paletteOffset := int64(offset + size - (3 * 256) - 2)
	if _, err := r.Seek(paletteOffset, io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek to palette data offset 0x%x: %w", paletteOffset, err)
	}

	return nil
}
