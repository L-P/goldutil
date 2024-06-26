package goldsrc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

const (
	TextureNameLen = 16

	// Fun fact: The letters MIP in the name are an acronym of the Latin phrase
	// multum in parvo, meaning "much in little". - Wikipedia.org.
	MIPLevelCount = 4
)

type TextureLump struct {
	Count uint32
	// Type signs are inconsistents in the documentation (VDN).
	Offsets  []int32 // len = Count, offset from TextureLump start
	Textures []TextureLumpEntry
}

type TextureName [TextureNameLen]byte

func (name TextureName) String() string {
	return string(name[0:bytes.IndexByte(name[:], 0)])
}

type TextureLumpEntry struct {
	Name          TextureName
	Width, Height uint32
	MIPOffsets    [MIPLevelCount]uint32
}

func (lump *TextureLump) Load(r io.ReadSeeker, entry LumpIndexEntry) error {
	if _, err := r.Seek(int64(entry.Offset), io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek to TextureLump start: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &lump.Count); err != nil {
		return fmt.Errorf("unable to read TextureLump.Count: %w", err)
	}
	if int(lump.Count)*int(unsafe.Sizeof(TextureLumpEntry{})) > int(entry.Length) {
		return fmt.Errorf("texture count %d exceeds available lump space %d", lump.Count, entry.Length)
	}

	if err := lump.loadHeader(r, entry); err != nil {
		return fmt.Errorf("unable to parse and load TextureLump header: %w", err)
	}

	if err := lump.loadTextures(r, entry); err != nil {
		return fmt.Errorf("unable to parse and load TextureLump textures: %w", err)
	}

	return nil
}

// FIXME: The {u,int32}ness of it all makes it very not correct.
func (lump *TextureLump) loadHeader(r io.ReadSeeker, entry LumpIndexEntry) error {
	var minTextureOffset = int32(4 + lump.Count*4) // Count + Offsets
	lump.Offsets = make([]int32, lump.Count)
	for i := 0; i < int(lump.Count); i++ {
		if err := binary.Read(r, binary.LittleEndian, &lump.Offsets[i]); err != nil {
			return fmt.Errorf("unable to read texture #%d offset: %w", i, err)
		}
		if lump.Offsets[i] < minTextureOffset || lump.Offsets[i] > entry.Length {
			return fmt.Errorf("texture #%d offset out of bound: %d < %d <", i, minTextureOffset, entry.Length)
		}
	}

	return nil
}

func (lump *TextureLump) loadTextures(r io.ReadSeeker, entry LumpIndexEntry) error {
	lump.Textures = make([]TextureLumpEntry, lump.Count)
	for i, offset := range lump.Offsets {
		if _, err := r.Seek(int64(entry.Offset+offset), io.SeekStart); err != nil {
			return fmt.Errorf("unable to seek to texture #%d start: %w", i, err)
		}

		if err := binary.Read(r, binary.LittleEndian, &lump.Textures[i]); err != nil {
			return fmt.Errorf("unable to read texture #%d: %w", i, err)
		}
	}

	return nil
}

func (lump *TextureLump) Validate() error {
	var errs []error

	for i, v := range lump.Textures {
		if v.Name[0] == 0 {
			errs = append(errs, fmt.Errorf("texture %d: empty texture name", i))
		}

		if n := bytes.IndexByte(v.Name[:], 0); n < 0 {
			errs = append(errs, fmt.Errorf("texture %d: no NUL", i))
		}

		if v.Width%16 != 0 {
			errs = append(errs, fmt.Errorf("texture %d (%s): width is not a multiple of 16: %d", i, v.Name.String(), v.Width))
		}
		if v.Height%16 != 0 {
			errs = append(errs, fmt.Errorf("texture %d (%s): height is not a multiple of 16: %d", i, v.Name.String(), v.Height))
		}
	}

	return errors.Join(errs...)
}
