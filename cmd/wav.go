package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
)

// https://developer.valvesoftware.com/wiki/Looping_a_sound
// The initial idea was to use go-audio/wav but it corrupts metadata and I did
// not want to write a RIFF/WAV parser.
var cueChunk = []byte{
	0x63, 0x75, 0x65, 0x20, // chunk ID, "cue "
	0x1C, 0x00, 0x00, 0x00, // size of the chunk: (12 + 24) - 8 = 28
	0x01, 0x00, 0x00, 0x00, // number of data points: 1
	0x01, 0x00, 0x00, 0x00, // ID of data point: 1
	0x00, 0x00, 0x00, 0x00, // position: If there is no playlist chunk, this is zero
	0x64, 0x61, 0x74, 0x61, // data chunk ID: "data"
	0x00, 0x00, 0x00, // chunk start: 0
	0x00, 0x00, 0x00, // block start: 0
	0x00, 0x00, 0x00, /// sample start: 0
}

func doWAVLoop(ctx context.Context, cmd *cli.Command) error {
	f, err := os.Open(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to open input WAV for reading: %w", err)
	}
	defer f.Close() //nolint:errcheck // readonly

	out, err := os.Create(cmd.String("out"))
	if err != nil {
		return fmt.Errorf("unable to create output file: %w", err)
	}

	if _, err := io.Copy(out, f); err != nil {
		return fmt.Errorf("unable to copy input WAV to output file: %w", err)
	}

	if _, err := out.Write(cueChunk); err != nil {
		return fmt.Errorf("unable to write CUE chunk: %w", err)
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("unable to finalize writing to output file: %w", err)
	}

	return nil
}
