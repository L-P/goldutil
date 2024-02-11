package wad

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"goldutil/set"
	"io"
	"os"
	"strings"
	"unsafe"
)

// Texture names are strings of 16 chars with a null terminator.
const (
	MaxNameLen = 15
	NameSize   = MaxNameLen + 1
)

// Fixed-size and \0-terminated string.
type TextureName [NameSize]byte

func (n TextureName) String() string {
	nul := bytes.IndexByte(n[:], 0)
	return string(n[:nul])
}

type WAD struct {
	Header

	textures           []texture
	nameToTextureIndex map[string]int
}

type texture struct {
	entry Entry
	mip   MIPTexture
}

// Returns available texture names.
// The canonical name is the directory entry name (all uppercase in halflife.wad).
// The texture lump name is unused.
func (wad WAD) Names() []string {
	names := make([]string, 0, len(wad.textures))
	for i := range wad.textures {
		names = append(names, wad.textures[i].entry.Name.String())
	}

	return names
}

func (wad WAD) GetTexture(name string) (MIPTexture, bool) {
	index, ok := wad.nameToTextureIndex[name]
	if !ok {
		return MIPTexture{}, false
	}

	return wad.textures[index].mip, true
}

func (wad WAD) String() string {
	var w strings.Builder

	w.WriteString("Header:\n")
	w.WriteString(wad.Header.String())

	fmt.Fprintf(&w, "\nDirectory (%d entries):", wad.Header.EntriesCount)
	for i, tex := range wad.textures {
		fmt.Fprintf(&w, "\nEntry #%d header:\n", i)
		w.WriteString(tex.entry.String())
		fmt.Fprintf(&w, "\nEntry #%d data:\n", i)
		w.WriteString(tex.mip.String())
	}

	return w.String()
}

// Binary-accurate.
type Header struct {
	MagicString [4]byte // "WAD3"

	// Directory
	EntriesCount  int32
	EntriesOffset int32 // from WAD start
}

func (wh Header) String() string {
	var w strings.Builder

	fmt.Fprintf(&w, "  MagicString: %s\n", wh.MagicString)
	fmt.Fprintf(&w, "  EntriesCount: %d\n", wh.EntriesCount)
	fmt.Fprintf(&w, "  EntriesOffset: 0x%x\n", wh.EntriesOffset)

	return w.String()
}

type EntryType byte

const EntryTypeMIPTex EntryType = 0x43

func (t EntryType) String() string {
	switch t {
	case EntryTypeMIPTex:
		return fmt.Sprintf("MIPTex (0x%X)", byte(t))
	default:
		return fmt.Sprintf("unknown (0x%X)", byte(t))
	}
}

const (
	EntrySize   = int32(unsafe.Sizeof(Entry{}))
	PaletteSize = int32(unsafe.Sizeof(Palette{}))
)

// Binary-accurate.
type Entry struct {
	Offset           int32 // offset to corresponding data (MIPTexture) from WAD start
	Size             int32
	UncompressedSize int32 // alway == Size (textures are never compressed)
	Type             EntryType
	Compressed       byte // always 0
	_                [2]byte
	Name             TextureName
}

func (e Entry) String() string {
	var w strings.Builder
	fmt.Fprintf(&w, "  Name: %s\n", e.Name)
	fmt.Fprintf(&w, "  Offset: 0x%x\n", e.Offset)
	fmt.Fprintf(&w, "  Size: %d\n", e.Size)
	fmt.Fprintf(&w, "  UncompressedSize: %d\n", e.UncompressedSize)
	fmt.Fprintf(&w, "  Type: %s\n", e.Type.String())
	fmt.Fprintf(&w, "  Compressed: %d\n", e.Compressed)
	return w.String()
}

func (wad *WAD) Read(r io.ReadSeeker) error {
	if err := binary.Read(r, binary.LittleEndian, &wad.Header); err != nil {
		return err
	}

	var WAD3 = [4]byte{'W', 'A', 'D', '3'}
	if wad.MagicString != WAD3 {
		return errors.New("cannot find magic string, probably not a WAD3 file")
	}

	var err error
	wad.textures, err = readTextures(r, wad.Header)
	if err != nil {
		return fmt.Errorf("unable to parse directory: %w", err)
	}

	wad.nameToTextureIndex = make(map[string]int, wad.EntriesCount)
	for i := range wad.textures {
		wad.nameToTextureIndex[wad.textures[i].entry.Name.String()] = i
	}

	return nil
}

// Combined directory / texture data reader.
func readTextures(r io.ReadSeeker, header Header) ([]texture, error) {
	var (
		ret   = make([]texture, 0, header.EntriesCount)
		names = set.NewPresenceSet[string](int(header.EntriesCount))
	)

	for i := int32(0); i < header.EntriesCount; i++ {
		offset := header.EntriesOffset + (EntrySize * i)
		if _, err := r.Seek(int64(offset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("unable to seek to offset %x of dir entry #%d", offset, i)
		}

		var entry Entry
		if err := binary.Read(r, binary.LittleEndian, &entry); err != nil {
			return nil, fmt.Errorf("unable to read entry #%d: %w", i, err)
		}

		entryName := entry.Name.String()
		if names.Has(entryName) {
			return nil, fmt.Errorf("entry #%d has a duplicated name: %s", i, entryName)
		}
		names.Set(entryName)

		if entry.Type != EntryTypeMIPTex {
			return nil, fmt.Errorf("unhandled entry #%d type: 0x%x", i, entry.Type)
		}

		var mip MIPTexture
		if err := mip.Read(r, entry.Offset, entry.Size); err != nil {
			return nil, fmt.Errorf("unable to read entry #%d MIPTexture data: %w", i, err)
		}

		ret = append(ret, texture{
			entry: entry,
			mip:   mip,
		})
	}

	return ret, nil
}

func NewFromFile(path string) (WAD, error) {
	f, err := os.Open(path)
	if err != nil {
		return WAD{}, fmt.Errorf("unable to open file: %w", err)
	}
	defer f.Close()

	var wad WAD
	if err := wad.Read(f); err != nil {
		return WAD{}, err
	}

	return wad, nil
}
