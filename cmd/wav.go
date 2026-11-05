package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

// The correct CUE chunk (36 bytes) according to the Microsoft WAV spec and
// Valve Developer Wiki (https://developer.valvesoftware.com/wiki/Looping_a_sound).
// The three bytes for chunk start, block start, and sample start were on
// vacation and are now back to work – making the chunk fully 36 bytes again.
// Without them, Source/GoldSrc engines won't recognise the loop point.
var cueChunk = []byte{
	0x63, 0x75, 0x65, 0x20, // "cue "
	0x1C, 0x00, 0x00, 0x00, // chunk size = 28
	0x01, 0x00, 0x00, 0x00, // number of cue points = 1
	0x01, 0x00, 0x00, 0x00, // cue point ID = 1
	0x00, 0x00, 0x00, 0x00, // position (zero if no playlist chunk)
	0x64, 0x61, 0x74, 0x61, // "data" chunk ID where the cue point lies
	0x00, 0x00, 0x00, 0x00, // chunk start = 0
	0x00, 0x00, 0x00, 0x00, // block start = 0
	0x00, 0x00, 0x00, 0x00, // sample start = 0 (loop from beginning)
}

func doWAVLoop(ctx context.Context, cmd *cli.Command) error {
	input := cmd.Args().First()
	if input == "" {
		return fmt.Errorf("no input file specified")
	}

	output := cmd.String("out")
	if output == "" {
		ext := filepath.Ext(input)
		base := strings.TrimSuffix(input, ext)
		output = base + "_looped" + ext
	}

	// Parse the WAV and strip all existing cue chunks
	chunks, _, err := parseWAV(input)
	if err != nil {
		return fmt.Errorf("failed to parse WAV: %w", err)
	}

	// Write a new WAV with cleaned chunks + our fresh cue chunk
	if err := writeWAV(output, chunks, cueChunk); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("Successfully written %s (loop point added)\n", output)
	return nil
}

// parseWAV reads a WAV file and returns a list of all chunks, excluding any
// "cue " chunks. It also returns the original RIFF size (unused but kept for
// potential future checks).
func parseWAV(filename string) ([][]byte, uint32, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	var header [12]byte
	if _, err := io.ReadFull(f, header[:]); err != nil {
		return nil, 0, fmt.Errorf("reading RIFF header: %w", err)
	}
	if string(header[0:4]) != "RIFF" {
		return nil, 0, fmt.Errorf("not a RIFF file")
	}
	if string(header[8:12]) != "WAVE" {
		return nil, 0, fmt.Errorf("not a WAVE file")
	}
	riffSize := binary.LittleEndian.Uint32(header[4:8])

	var chunks [][]byte
	for {
		var chunkHeader [8]byte
		_, err := io.ReadFull(f, chunkHeader[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("reading chunk header: %w", err)
		}
		chunkID := string(chunkHeader[0:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		data := make([]byte, chunkSize)
		if _, err := io.ReadFull(f, data); err != nil {
			return nil, 0, fmt.Errorf("reading chunk data (ID %s): %w", chunkID, err)
		}

		// Skip cue chunks – we'll add our own later
		if chunkID == "cue " {
			continue
		}

		// Store the chunk as a complete block: ID (4) + size (4) + data
		chunk := make([]byte, 8+chunkSize)
		copy(chunk[0:4], chunkHeader[0:4])
		binary.LittleEndian.PutUint32(chunk[4:8], chunkSize)
		copy(chunk[8:], data)
		chunks = append(chunks, chunk)
	}

	return chunks, riffSize, nil
}

// writeWAV writes a new WAV file from the given list of chunks and appends
// the supplied cue chunk at the end. It recalculates the RIFF size correctly.
func writeWAV(filename string, chunks [][]byte, newCueChunk []byte) error {
	// Calculate total size of the data after "RIFF" + size field.
	// The "WAVE" identifier (4 bytes) plus the sum of all chunk lengths.
	var totalDataSize uint32 = 4 // "WAVE"
	for _, ch := range chunks {
		totalDataSize += uint32(len(ch))
	}
	totalDataSize += uint32(len(newCueChunk))

	// In a RIFF file, the size field = total file size - 8 bytes.
	newRiffSize := totalDataSize

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write "RIFF"
	if _, err := out.Write([]byte("RIFF")); err != nil {
		return err
	}
	// Write size (little‑endian)
	if err := binary.Write(out, binary.LittleEndian, newRiffSize); err != nil {
		return err
	}
	// Write "WAVE"
	if _, err := out.Write([]byte("WAVE")); err != nil {
		return err
	}
	// Write all stored chunks
	for _, ch := range chunks {
		if _, err := out.Write(ch); err != nil {
			return err
		}
	}
	// Append our brand new, perfectly sized cue chunk
	if _, err := out.Write(newCueChunk); err != nil {
		return err
	}

	return nil
}