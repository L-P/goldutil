package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"

	"github.com/L-P/goldutil/goldsrc"
	"github.com/L-P/goldutil/goldsrc/bsp"
)

func doBSPRemapMaterials(ctx context.Context, cmd *cli.Command) error {
	source, err := goldsrc.LoadMaterialsFromFile(cmd.String("original-materials"))
	if err != nil {
		return fmt.Errorf("unable to load original-materials: %w", err)
	}

	replacement, err := goldsrc.LoadMaterialsFromFile(cmd.String("replacement-materials"))
	if err != nil {
		return fmt.Errorf("unable to load replacement-materials: %w", err)
	}

	if source.IsEmpty() || replacement.IsEmpty() {
		return errors.New("no materials in source or replacement list")
	}

	bsp, err := bsp.LoadFromFile(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to load BSP: %w", err)
	}

	var (
		verbose  = cmd.Bool("verbose")
		remapper = goldsrc.NewMaterialsRemapper(source)
	)
	mapping, err := remapper.ReMap(cmd.ErrWriter, bsp.Textures.Textures, replacement)
	if err != nil {
		return fmt.Errorf("unable to remap materials: %w", err)
	}

	for i, tex := range bsp.Textures.Textures {
		mapTo, ok := mapping[tex.Name]
		if !ok {
			continue
		}

		if verbose {
			fmt.Fprintf(
				cmd.Writer,
				"Remapping %-15s to %s\n",
				strings.ToUpper(tex.Name.String()),
				strings.ToUpper(mapTo.String()),
			)
		}

		bsp.Textures.Textures[i].Name = mapTo
	}

	if err := bsp.WriteToFile(cmd.String("out")); err != nil {
		return fmt.Errorf("unable to write BSP: %w", err)
	}

	if cmd.Bool("verbose") {
		remapper.PrintAvailable(cmd.Writer)
	}

	return nil
}

func doBSPInfo(ctx context.Context, cmd *cli.Command) error {
	bsp, err := bsp.LoadFromFile(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to load BSP: %w", err)
	}

	fmt.Fprint(cmd.Writer, bsp.String())

	return nil
}

func doBSPLimits(ctx context.Context, cmd *cli.Command) error {
	bsp, err := bsp.LoadFromFile(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to load BSP: %w", err)
	}

	fmt.Fprintf(
		cmd.Writer,
		"%-18s % 9s % 9s % 4s\n",
		"Type", "Current", "Max", "Pct",
	)

	yellow := color.New(color.FgYellow).Fprintf
	red := color.New(color.FgRed).Fprintf
	var errs []error

	for _, v := range bsp.Limits() {
		if v.Max <= 0 {
			return fmt.Errorf("developer error, invalid limit for: %s", v.Desc)
		}

		pct := math.Ceil(float64(v.Current) / float64(v.Max) * 100)
		var printer = fmt.Fprintf
		if pct > 60 {
			printer = yellow
		}
		if pct > 80 {
			printer = red
		}

		//nolint:errcheck
		printer(
			cmd.Writer,
			"%-18s % 9d % 9d % 3.0f%%\n",
			v.Desc, v.Current, v.Max, pct,
		)

		if pct > 100 {
			errs = append(errs, fmt.Errorf("exceeded limit on %s", v.Desc))
		}
	}

	return errors.Join(errs...)
}
