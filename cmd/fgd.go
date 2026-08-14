package main

import (
	"context"
	"fmt"
	"os"

	"github.com/L-P/goldutil/fgd"
	"github.com/urfave/cli/v3"
)

func doFGDCheck(ctx context.Context, cmd *cli.Command) error {
	_, err := loadFGD(cmd)
	if err != nil {
		return fmt.Errorf("unable to load FGD: %w", err)
	}

	return nil
}

func loadFGD(cmd *cli.Command) (fgd.FGD, error) {
	if path := cmd.Args().First(); path != "" {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("unable to open FGD file for reading: %w", err)
		}
		defer f.Close() //nolint: errcheck // read-only

		return fgd.NewFromReader(f)
	}

	return fgd.NewFromReader(cmd.Reader)
}
