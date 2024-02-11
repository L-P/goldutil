package main

import (
	"errors"
	"flag"
	"fmt"
	"goldutil/wad"
	"os"
)

func doWADExtract(args []string) error {
	fset := flag.NewFlagSet("wad-extract", flag.ExitOnError)
	fset.Usage = usage
	dir := fset.String("out", "", "destination directory")
	if err := fset.Parse(args); err != nil {
		return err
	}

	stat, err := os.Stat(*dir)
	if err != nil {
		return fmt.Errorf("unable to use destination directory: %w", err)
	}
	if err == nil && !stat.IsDir() {
		return errors.New("output directory paths exists but is not a directory")
	}

	wad3, err := wad.NewFromFile(fset.Arg(0))
	if err != nil {
		return fmt.Errorf("unable to open and parse WAD file: %w", err)
	}

	fmt.Println(wad3.String())

	return nil
}

func doWADInfo(args []string) error {
	fset := flag.NewFlagSet("wad-info", flag.ExitOnError)
	fset.Usage = usage
	if err := fset.Parse(args); err != nil {
		return err
	}

	wad3, err := wad.NewFromFile(fset.Arg(0))
	if err != nil {
		return fmt.Errorf("unable to open and parse WAD file: %w", err)
	}

	fmt.Println(wad3.String())

	return nil
}

func doWADCreate(args []string) error {
	return errors.New("not implemented")
}
