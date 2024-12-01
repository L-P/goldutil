package wad

import (
	"encoding/binary"
	"errors"
	"fmt"
	"goldutil/sprite"
	"io"
	"strings"
)

// Number of mimmaps per texture, base texture is mipmap 0.
const NumMIPMaps = 4

// Binary-accurate.
type MIPTextureHeader struct {
	// Seems to be a lowercase version of the dir entry name. Unused?
	Name TextureName

	// multiples of 16 (guarantees 4 half-size mipmaps)
	Width  int32
	Height int32

	MIPOffsets [NumMIPMaps]int32
}

func (h MIPTextureHeader) IsEmbedded() bool {
	return h.MIPOffsets[0] > 0
}

type MIPTexture struct {
	MIPTextureHeader

	// total len = w*h + (w*h)/2 + (w*h)/4 + (w*h)/8 = (15*(w*h))/8
	// Paletted data, 1bpp.
	MIPData     [NumMIPMaps][]byte
	PaletteSize int16          // always 256
	Palette     sprite.Palette // TODO move palette out of sprite
	_           [2]byte
}

func (mip MIPTexture) Size() int32 {
	// 2 bytes of padding, 2 bytes of palette size, the palette, the header, and the data
	var ret = 2 + 2 + PaletteDataSize + MIPTextureHeaderSize
	for i := range mip.MIPData {
		ret += int32(len(mip.MIPData[i]))
	}

	return ret
}

func NewMIPTexture(nameStr string, width, height int) (MIPTexture, error) {
	// Lumps names are lowercase and there's hardcoded case-sensitive checks
	// on some texture names (eg. "sky").
	name, err := NewTextureName(strings.ToLower(nameStr))
	if err != nil {
		return MIPTexture{}, fmt.Errorf("unable to create texture name: %w", err)
	}

	if width%16 != 0 || height%16 != 0 {
		return MIPTexture{}, errors.New("dimensions not divisible by 16")
	}

	return MIPTexture{
		PaletteSize: 256,
		MIPTextureHeader: MIPTextureHeader{
			Name:   name,
			Width:  int32(width),
			Height: int32(height),
		},
	}, nil
}

func (mip MIPTexture) String() string {
	var w strings.Builder
	fmt.Fprintf(&w, "  Name: %s\n", mip.Name.String())
	fmt.Fprintf(&w, "  Width: %d\n", mip.Width)
	fmt.Fprintf(&w, "  Height: %d\n", mip.Height)
	fmt.Fprintf(&w, "  Embedded: %t\n", mip.IsEmbedded())
	if !mip.IsEmbedded() {
		return w.String()
	}

	fmt.Fprintf(&w, "  PaletteSize: %d\n", mip.PaletteSize)

	for i := range mip.MIPOffsets {
		fmt.Fprintf(&w, "  MIPMap #%d Offset: 0x%x\n", i, mip.MIPOffsets[i])
		fmt.Fprintf(&w, "  MIPMap #%d Size: %d\n", i, len(mip.MIPData[i]))
	}

	return w.String()
}

func (mip *MIPTexture) Read(r io.ReadSeeker, offset int32) error {
	if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek to offset %x of MIPTexture header", offset)
	}

	if err := binary.Read(r, binary.LittleEndian, &mip.MIPTextureHeader); err != nil {
		return fmt.Errorf("unable to read MIPTextureHeader: %w", err)
	}

	if !mip.IsEmbedded() {
		return nil
	}

	for i := range mip.MIPData {
		// Each MIPMap is half the size of the previous one.
		scale := int32(mipIndexToScale(i))
		pixSize := (mip.Width / scale) * (mip.Height / scale)
		mip.MIPData[i] = make([]byte, pixSize)

		mipmapOffset := int64(offset + mip.MIPOffsets[i])
		if _, err := r.Seek(mipmapOffset, io.SeekStart); err != nil {
			return fmt.Errorf("unable to seek to MIPMap data #%d: %w", i, err)
		}

		if err := binary.Read(r, binary.LittleEndian, &mip.MIPData[i]); err != nil {
			return fmt.Errorf("unable to read MIPMap #%d: %w", i, err)
		}
	}

	if err := binary.Read(r, binary.LittleEndian, &mip.PaletteSize); err != nil {
		return fmt.Errorf("unable to read PaletteSize: %w", err)
	}

	paletteOffset := int64(offset + mip.Size() - PaletteDataSize - 2)
	if _, err := r.Seek(paletteOffset, io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek to palette data offset 0x%x: %w", paletteOffset, err)
	}

	if err := binary.Read(r, binary.LittleEndian, &mip.Palette); err != nil {
		return fmt.Errorf("unable to read Palette: %w", err)
	}

	return nil
}

func (mip *MIPTexture) SetData(pix []byte) error {
	if len(pix) != int(mip.Width*mip.Height) {
		return errors.New("data length doesn't match texture size")
	}

	mip.MIPData[0] = make([]byte, len(pix))
	copy(mip.MIPData[0], pix)

	for mipID := 1; mipID < NumMIPMaps; mipID++ {
		scale := int32(mipIndexToScale(mipID))
		w, h := (mip.Width / scale), (mip.Height / scale)

		mip.MIPData[mipID] = make([]byte, w*h)
		for i := range mip.MIPData[mipID] {
			x, y := int32(i)%w, int32(i)/w
			src := ((y * scale) * mip.Width) + (x * scale)
			mip.MIPData[mipID][i] = mip.MIPData[0][src]
		}
	}

	var offset = MIPTextureHeaderSize
	for i := range mip.MIPData {
		mip.MIPOffsets[i] = offset
		offset += int32(len(mip.MIPData[i]))
	}

	return nil
}

func (mip *MIPTexture) Write(w io.Writer) (int, error) {
	if err := binary.Write(w, binary.LittleEndian, mip.MIPTextureHeader); err != nil {
		return 0, fmt.Errorf("unable to write MIPTextureHeader: %w", err)
	}

	for mipID := range mip.MIPData {
		if err := binary.Write(w, binary.LittleEndian, mip.MIPData[mipID]); err != nil {
			return 0, fmt.Errorf("unable to write mip data #%d: %w", mipID, err)
		}
	}

	if err := binary.Write(w, binary.LittleEndian, mip.PaletteSize); err != nil {
		return 0, fmt.Errorf("unable to write PaletteSize in mip: %w", err)
	}

	if err := binary.Write(w, binary.LittleEndian, mip.Palette); err != nil {
		return 0, fmt.Errorf("unable to write palette: %w", err)
	}

	if err := binary.Write(w, binary.LittleEndian, [2]byte{}); err != nil {
		return 0, fmt.Errorf("unable to write padding: %w", err)
	}

	return int(mip.Size()), nil
}
