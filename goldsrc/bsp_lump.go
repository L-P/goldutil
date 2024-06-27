package goldsrc

import (
	"fmt"
	"io"
)

type LumpType int

const ( // DO NOT SORT
	LumpTypeEntities LumpType = iota
	LumpTypePlanes
	LumpTypeTextures
	LumpTypeVertices
	LumpTypeVisibility
	LumpTypeNodes
	LumpTypeTexInfo
	LumpTypeFaces
	LumpTypeLighting
	LumpTypeClipNodes
	LumpTypeLeaves
	LumpTypeMarkSurfaces
	LumpTypeEdges
	LumpTypeSurfEdges
	LumpTypeModels

	LumpIndexSize
)

func (t LumpType) String() string {
	switch t {
	case LumpTypeEntities:
		return "LumpTypeEntities"
	case LumpTypePlanes:
		return "LumpTypePlanes"
	case LumpTypeTextures:
		return "LumpTypeTextures"
	case LumpTypeVertices:
		return "LumpTypeVertices"
	case LumpTypeVisibility:
		return "LumpTypeVisibility"
	case LumpTypeNodes:
		return "LumpTypeNodes"
	case LumpTypeTexInfo:
		return "LumpTypeTexInfo"
	case LumpTypeFaces:
		return "LumpTypeFaces"
	case LumpTypeLighting:
		return "LumpTypeLighting"
	case LumpTypeClipNodes:
		return "LumpTypeClipNodes"
	case LumpTypeLeaves:
		return "LumpTypeLeaves"
	case LumpTypeMarkSurfaces:
		return "LumpTypeMarkSurfaces"
	case LumpTypeEdges:
		return "LumpTypeEdges"
	case LumpTypeSurfEdges:
		return "LumpTypeSurfEdges"
	case LumpTypeModels:
		return "LumpTypeModels"
	case LumpIndexSize: // fail below
	}

	return fmt.Sprintf("<invalid: %d>", t)
}

func init() {
	if LumpIndexSize != 15 {
		panic("LumpIndexSize != 15")
	}
}

type LumpIndexEntry struct {
	Offset, Length int32
}

type Lump interface {
	Load(io.ReadSeeker, LumpIndexEntry) error
	Write(w io.WriteSeeker) (int, error)
	Validate() error
	String() string
}
type RawLump []byte

func (lump *RawLump) String() string {
	return " n/a\n"
}

func (lump *RawLump) Load(r io.ReadSeeker, entry LumpIndexEntry) error {
	*lump = make([]byte, entry.Length)
	if _, err := r.Seek(int64(entry.Offset), io.SeekStart); err != nil {
		return fmt.Errorf("unable to seek to %d: %w", entry.Offset, err)
	}

	if n, err := io.ReadFull(r, []byte(*lump)); err != nil {
		return fmt.Errorf("count only read %d out of %d bytes: %w", n, entry.Length, err)
	}

	return nil
}

func (lump *RawLump) Write(w io.WriteSeeker) (int, error) {
	return w.Write(*lump)
}

func (lump *RawLump) Validate() error {
	return nil
}
