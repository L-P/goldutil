package goldsrc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"goldutil/set"
	"goldutil/wad"
	"io"
	"strings"
)

type TextureLump struct {
	Count uint32
	// Type signs are inconsistents in the documentation (VDN).
	Offsets  []int32 // len = Count, offset from TextureLump start
	Textures []wad.MIPTexture
}

func (lump *TextureLump) Load(r io.ReadSeeker, entry LumpIndexEntry) error {
	if _, err := r.Seek(int64(entry.Offset), io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek to TextureLump start: %w", err)
	}

	if err := binary.Read(r, binary.LittleEndian, &lump.Count); err != nil {
		return fmt.Errorf("unable to read TextureLump.Count: %w", err)
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
			return fmt.Errorf(
				"texture #%d offset out of bounds: %d < %d < %d",
				i,
				minTextureOffset, lump.Offsets[i], minTextureOffset+entry.Length,
			)
		}
	}

	return nil
}

func (lump *TextureLump) loadTextures(r io.ReadSeeker, entry LumpIndexEntry) error {
	lump.Textures = make([]wad.MIPTexture, lump.Count)
	for i, offset := range lump.Offsets {
		if err := lump.Textures[i].Read(r, entry.Offset+offset); err != nil {
			return fmt.Errorf("unable to read texture #%d: %w", i, err)
		}
	}

	return nil
}

func (lump *TextureLump) Validate() error {
	var (
		errs []error
		seen = set.NewPresenceSet[string](len(lump.Textures))
	)

	for i, v := range lump.Textures {
		if v.Name[0] == 0 {
			errs = append(errs, fmt.Errorf("texture %d: empty texture name", i))
		}
		if n := bytes.IndexByte(v.Name[:], 0); n < 0 {
			errs = append(errs, fmt.Errorf("texture %d: no NUL", i))
		}
		var name = v.Name.String()
		if !isValidTextureName(strings.ToUpper(name)) {
			errs = append(errs, fmt.Errorf("texture %d: invalid chars in name", i))
		}

		if v.Width%16 != 0 {
			errs = append(errs, fmt.Errorf("texture %d (%s): width is not a multiple of 16: %d", i, name, v.Width))
		}
		if v.Height%16 != 0 {
			errs = append(errs, fmt.Errorf("texture %d (%s): height is not a multiple of 16: %d", i, name, v.Height))
		}

		var zeroes int
		for _, v := range v.MIPOffsets {
			if v == 0 {
				zeroes++
			}
		}
		if zeroes != 0 && zeroes != len(v.MIPOffsets) {
			errs = append(errs, fmt.Errorf("texture %d (%s): missing some mimap offsets", i, name))
		}

		if seen.Has(name) {
			errs = append(errs, fmt.Errorf("texture %d (%s): duplicate texture name", i, name))
		}
		seen.Set(name)
	}

	return errors.Join(errs...)
}

func (lump *TextureLump) Write(w io.WriteSeeker) (int, error) {
	var offset int

	if err := binary.Write(w, binary.LittleEndian, lump.Count); err != nil {
		return offset, fmt.Errorf("unable to write texture count: %w", err)
	}
	offset += binary.Size(lump.Count)

	var offsetsOffset = offset // semantic satiety anyone?
	if err := binary.Write(w, binary.LittleEndian, lump.Offsets); err != nil {
		return offset, fmt.Errorf("unable to write provisional texture offsets: %w", err)
	}
	offset += binary.Size(lump.Offsets)

	for i, v := range lump.Textures {
		n, err := v.Write(w)
		if err != nil {
			return offset, fmt.Errorf("unable to write texture #%d: %w", i, err)
		}
		offset += n
	}

	if offset < offsetsOffset {
		panic("unreachable, offset < offsetsOffset")
	}

	if _, err := w.Seek(-int64(offset-offsetsOffset), io.SeekCurrent); err != nil {
		return offset, fmt.Errorf("unable to seek to start of file to finalize header: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, lump.Offsets); err != nil {
		return offset, fmt.Errorf("unable to write provisional texture offsets: %w", err)
	}

	return offset, nil
}

func (lump *TextureLump) String() string {
	var b strings.Builder
	b.WriteRune('\n')

	fmt.Fprintf(&b, "  Count: %d\n", lump.Count)
	b.WriteString("  Offsets: ")
	for _, v := range lump.Offsets {
		fmt.Fprintf(&b, "0x%08x ", v)
	}
	b.WriteRune('\n')

	for _, v := range lump.Textures {
		b.WriteString(v.String())
	}

	return b.String()

}
