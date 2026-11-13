package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/L-P/goldutil/goldsrc/qmap"
	"github.com/L-P/goldutil/neat"
)

func doMapGraph(ctx context.Context, cmd *cli.Command) error {
	qm, err := loadQMap(cmd)
	if err != nil {
		return fmt.Errorf("unable to read from map: %w", err)
	}

	GraphQMap(qm, os.Stdout)

	return nil
}

func doNeat(ctx context.Context, cmd *cli.Command) error {
	qm, err := loadQMap(cmd)
	if err != nil {
		return fmt.Errorf("unable to read from map: %w", err)
	}

	mod, err := os.OpenRoot(cmd.String("moddir"))
	if err != nil {
		return fmt.Errorf("unable to open current working directory: %w", err)
	}

	if err := neat.Neatify(qm, mod); err != nil {
		return fmt.Errorf("unable to neatify map: %w", err)
	}

	outPath := cmd.String("out")
	if outPath == "" {
		_, err = fmt.Fprint(cmd.Writer, qm.String())
		return err
	}

	dest, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("unable to open '%s' for writing: %w", outPath, err)
	}

	if _, err := fmt.Fprint(dest, qm.String()); err != nil {
		return fmt.Errorf("unable to write to map file: %w", err)
	}

	if err := dest.Close(); err != nil {
		return fmt.Errorf("unable to finalize writing to '%s': %w", outPath, err)
	}

	return nil
}

func doMapExport(ctx context.Context, cmd *cli.Command) error {
	qm, err := loadQMap(cmd)
	if err != nil {
		return fmt.Errorf("unable to read from map: %w", err)
	}

	clean, err := exportQMap(qm, cmd.Bool("cleanup-tb"), cmd.Bool("fix-zero-angles"))
	if err != nil {
		return fmt.Errorf("unable to export map: %w", err)
	}

	fmt.Fprint(cmd.Writer, clean.String())

	return nil
}

func loadQMap(cmd *cli.Command) (*qmap.QMap, error) {
	if path := cmd.Args().First(); path != "" {
		return qmap.LoadFromFile(path)
	}

	return qmap.LoadFromReader(cmd.Reader)
}
