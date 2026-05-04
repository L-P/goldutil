package bsp

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

// hard limits courtesy of SDHLT bspfile.h, -1 means infinite.
const (
	MaxMapClipNodes    = 0x7FFF
	MaxMapEdges        = 256000
	MaxMapEntities     = 16384
	MaxMapEntString    = 2048 * 1024 // raw string data, not parsed entities count
	MaxMapFaces        = 0xFFFF
	MaxMapLeaves       = 32760
	MaxMapLighting     = 0x3000000
	MaxMapMarkSurfaces = 0xFFFF
	MaxMapModels       = 512
	MaxMapNodes        = 0x7FFF
	MaxMapPlanes       = 0x7FFF
	MaxMapSurfEdges    = 512000
	MaxMapTexInfo      = 0x7FFF
	// VHLT says 4096 but it must be the _number_ of textures, not the lump size. I went over 100 MB and everything was fine.
	MaxMapTextures   = -1
	MaxMapVertices   = 0xFFFF
	MaxMapVisibility = 0x800000
)

func (t LumpType) EntrySize() int {
	switch t {
	case LumpTypeClipNodes:
		return 8
	case LumpTypeEdges:
		return 2
	case LumpTypeEntities:
		return 1
	case LumpTypeFaces:
		return 20
	case LumpTypeLeaves:
		return 28
	case LumpTypeLighting:
		return 1
	case LumpTypeMarkSurfaces:
		return 2
	case LumpTypeModels:
		return 64
	case LumpTypeNodes:
		return 24
	case LumpTypePlanes:
		return 20
	case LumpTypeSurfEdges:
		return 4
	case LumpTypeTexInfo:
		return 40
	case LumpTypeTextures:
		return 40
	case LumpTypeVertices:
		return 12
	case LumpTypeVisibility:
		return 1
	case LumpIndexSize: // fail below
	}

	panic("invalid LumpType")
}

func (t LumpType) Limit() int {
	switch t {
	case LumpTypeEntities:
		return MaxMapEntString
	case LumpTypePlanes:
		return MaxMapPlanes
	case LumpTypeTextures:
		return MaxMapTextures
	case LumpTypeVertices:
		return MaxMapVertices
	case LumpTypeVisibility:
		return MaxMapVisibility
	case LumpTypeNodes:
		return MaxMapNodes
	case LumpTypeTexInfo:
		return MaxMapTexInfo
	case LumpTypeFaces:
		return MaxMapFaces
	case LumpTypeLighting:
		return MaxMapLighting
	case LumpTypeClipNodes:
		return MaxMapClipNodes
	case LumpTypeLeaves:
		return MaxMapLeaves
	case LumpTypeMarkSurfaces:
		return MaxMapMarkSurfaces
	case LumpTypeEdges:
		return MaxMapEdges
	case LumpTypeSurfEdges:
		return MaxMapSurfEdges
	case LumpTypeModels:
		return MaxMapModels
	case LumpIndexSize: // fail below
	}

	panic("invalid LumpType")
}

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
