package main

import (
	"context"
	"errors"
	"fmt"
	"goldutil/goldsrc/qmap"
	"goldutil/neat"
	"os"

	"github.com/urfave/cli/v3"
)

func doMapGraph(ctx context.Context, cmd *cli.Command) error {
	path := cmd.Args().Get(0)
	if path == "" {
		return errors.New("expected one argument: the .map to parse and graph")
	}

	qm, err := loadQMap(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to read from map: %w", err)
	}

	GraphQMap(qm, os.Stdout)

	return nil
}

func doNeat(ctx context.Context, cmd *cli.Command) error {
	qm, err := loadQMap(cmd.Args().Get(0))
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

	fmt.Fprint(cmd.Writer, qm.String())

	return nil
}

func doMapExport(ctx context.Context, cmd *cli.Command) error {
	qm, err := loadQMap(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to read from map: %w", err)
	}

	clean, err := exportQMap(qm, cmd.Bool("cleanup-tb"))
	if err != nil {
		return fmt.Errorf("unable to export map: %w", err)
	}

	fmt.Fprint(cmd.Writer, clean.String())

	return nil
}

func loadQMap(path string) (*qmap.QMap, error) {
	if path == "" {
		return qmap.LoadFromReader(os.Stdin)
	}

	return qmap.LoadFromFile(path)
}
