package goldsrc

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

const BSPVersionGoldSrc = 30

// BSP holds a full BSP in memory.
type BSP struct {
	BSPHeader

	Entities     RawLump
	Planes       RawLump
	Textures     TextureLump
	Vertices     RawLump
	Visibility   RawLump
	Nodes        RawLump
	TexInfo      RawLump
	Faces        RawLump
	Lighting     RawLump
	ClipNodes    RawLump
	Leaves       RawLump
	MarkSurfaces RawLump
	Edges        RawLump
	SurfEdges    RawLump
	Models       RawLump
}

type BSPHeader struct {
	Version   int32
	LumpIndex [LumpIndexSize]LumpIndexEntry
}

func (h BSPHeader) Validate() error {
	if h.Version != BSPVersionGoldSrc {
		return fmt.Errorf("unable to read BSP version other than %d, got: %d", BSPVersionGoldSrc, h.Version)
	}

	var size = int32(binary.Size(h))
	for _, v := range h.LumpIndex {
		size += v.Length
	}

	for i, v := range h.LumpIndex {
		var typ = LumpType(i)

		if v.Offset > size {
			return fmt.Errorf("%s offset is out of bounds: %x > %x", typ.String(), v.Offset, size)
		}
		if v.Offset < 0 {
			return fmt.Errorf("%s offset is out of bounds: %x < 0", typ.String(), v.Offset)
		}
	}

	return nil
}

func (h BSPHeader) String() string {
	var b strings.Builder
	fmt.Fprintln(&b, "BSPHeader:")
	for i, v := range h.LumpIndex {
		typ := LumpType(i)
		fmt.Fprintf(
			&b,
			"  - %-21s[0x%08x;0x%08x] % 8s (%d)\n",
			typ.String(),
			v.Offset, v.Offset+v.Length,
			humanize(int(v.Length)),
			v.Length,
		)
		_ = v
	}

	return b.String()
}

func (bsp *BSP) String() string {
	var b strings.Builder
	b.WriteString(bsp.BSPHeader.String())

	for i, v := range bsp.Lumps() {
		typ := LumpType(i)
		fmt.Fprintf(&b, "%s:", typ.String())
		b.WriteString(v.String())
	}

	return b.String()
}

func (bsp *BSP) Lumps() []Lump { // MUST match LumpIndex order.
	return []Lump{
		&bsp.Entities,
		&bsp.Planes,
		&bsp.Textures,
		&bsp.Vertices,
		&bsp.Visibility,
		&bsp.Nodes,
		&bsp.TexInfo,
		&bsp.Faces,
		&bsp.Lighting,
		&bsp.ClipNodes,
		&bsp.Leaves,
		&bsp.MarkSurfaces,
		&bsp.Edges,
		&bsp.SurfEdges,
		&bsp.Models,
	}
}

func LoadBSP(r io.ReadSeeker) (*BSP, error) {
	var bsp BSP

	if err := binary.Read(r, binary.LittleEndian, &bsp.BSPHeader); err != nil {
		return nil, fmt.Errorf("unable to read header: %w", err)
	}
	if err := bsp.BSPHeader.Validate(); err != nil {
		return nil, fmt.Errorf("unable to validate header: %w", err)
	}

	for i, lump := range bsp.Lumps() {
		typ := LumpType(i)
		if err := lump.Load(r, bsp.BSPHeader.LumpIndex[i]); err != nil {
			return nil, fmt.Errorf("unable to load %s: %w", typ.String(), err)
		}

		if err := lump.Validate(); err != nil {
			return nil, fmt.Errorf("unable to validate %s: %w", typ.String(), err)
		}
	}

	return &bsp, nil
}

func LoadBSPFromFile(path string) (*BSP, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open BSP for reading: %w", err)
	}
	defer f.Close()

	return LoadBSP(f)
}

func humanize(bytes int) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "ZiB"}
	for i := len(units) - 1; i >= 0; i-- {
		d := 1 << (10 * i)
		if bytes >= d {
			return fmt.Sprintf("%.0f %s", float64(bytes)/float64(d), units[i])
		}
	}

	return "0 B"
}

func (bsp *BSP) WriteToFile(path string) error {
	out, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open file for writing: %w", err)
	}

	if err := bsp.Write(out); err != nil {
		out.Close()
		return fmt.Errorf("unable to write to BSP: %w", err)
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("unable to finish writing to BSP: %w", err)
	}

	return nil
}

func (bsp *BSP) Write(w io.WriteSeeker) error {
	if _, err := w.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek to start of file: %w", err)
	}

	bsp.BSPHeader = BSPHeader{Version: BSPVersionGoldSrc}
	if err := binary.Write(w, binary.LittleEndian, bsp.BSPHeader); err != nil {
		return fmt.Errorf("unable to write provisional BSP header: %w", err)
	}
	offset := binary.Size(bsp.BSPHeader)

	// HACK HACK HACK HACK HACK
	// Keep order as found in map compiled by ericw, right now we can only
	// patch textures, full BSP reimplementation is necessary to actually write
	// the file, there's offsets all around.
	order := []LumpType{
		LumpTypePlanes,
		LumpTypeLeaves,
		LumpTypeVertices,
		LumpTypeNodes,
		LumpTypeTexInfo,
		LumpTypeFaces,
		LumpTypeClipNodes,
		LumpTypeMarkSurfaces,
		LumpTypeSurfEdges,
		LumpTypeEdges,
		LumpTypeModels,
		LumpTypeLighting,
		LumpTypeVisibility,
		LumpTypeEntities,
		LumpTypeTextures,
	}

	for _, typ := range order {
		lump := bsp.Lumps()[typ]
		n, err := lump.Write(w)
		if err != nil {
			return fmt.Errorf("unable to write lump %s: %w", typ.String(), err)
		}

		bsp.BSPHeader.LumpIndex[typ] = LumpIndexEntry{
			Offset: int32(offset),
			Length: int32(n),
		}

		offset += n

		if delta := offset % 4; delta != 0 {
			var padding = make([]byte, 4-delta)
			if _, err := w.Write(padding); err != nil {
				return fmt.Errorf("unable to write padding: %w", err)
			}
			offset += len(padding)
		}
	}

	if _, err := w.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek to start of file to finalize header: %w", err)
	}
	if err := binary.Write(w, binary.LittleEndian, bsp.BSPHeader); err != nil {
		return fmt.Errorf("unable to write final BSP header: %w", err)
	}

	return nil
}
