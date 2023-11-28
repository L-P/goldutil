package main

import (
	"errors"
	"flag"
	"fmt"
	"goldutil/qmap"
	"goldutil/sprite"
	"os"
)

var help = `Usage: %s COMMAND [ARGS…]

Commands:
    entity-graph MAP
        Outputs a graphviz digraph of entity caller/callee relationships from a
        .map file. ripent exports use the same format and can be read too.

    sprite-info SPR
        Prints parsed frame data from a sprite.

    sprite-extract SPR [-dir DIR]
        Outputs all frames of a sprite to the current directory.

        Options:
            -dir DIR    Outputs frames to the specified directory instead of
                        the current one.
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), help, os.Args[0])
	}

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, help, os.Args[0])
		os.Exit(1)
	}

	if err := dispatch(os.Args[1], os.Args[2:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			flag.Usage()
			return
		}

		fmt.Fprintln(os.Stderr, err.Error())
	}
}

func dispatch(command string, args []string) error {
	switch command {
	case "entity-graph":
		return doEntGraph(args)
	case "sprite-extract":
		return doSpriteExtract(args)
	case "sprite-info":
		return doSpriteInfo(args)
	}

	return nil
}

func doSpriteExtract(args []string) error {
	fset := flag.NewFlagSet("sprite-extract", flag.ExitOnError)
	dir := fset.String("dir", "", "destination directory")

	if err := fset.Parse(args); err != nil {
		return err
	}

	path := fset.Arg(0)
	if path == "" {
		return errors.New("expected one argument: the .spr to parse and extract")
	}

	spr, err := sprite.NewFromFile(path)
	if err != nil {
		return fmt.Errorf("unable to open sprite: %w", err)
	}

	return extractSprite(spr, *dir)
}

func doSpriteInfo(args []string) error {
	fset := flag.NewFlagSet("sprite-info", flag.ExitOnError)
	if err := fset.Parse(args); err != nil {
		return err
	}

	path := fset.Arg(0)
	if path == "" {
		return errors.New("expected one argument: the .spr to parse and display")
	}

	spr, err := sprite.NewFromFile(path)
	if err != nil {
		return fmt.Errorf("unable to open sprite: %w", err)
	}

	fmt.Println(spr.String())

	return nil
}

func doEntGraph(args []string) error {
	fset := flag.NewFlagSet("entity-graph", flag.ExitOnError)
	if err := fset.Parse(args); err != nil {
		return err
	}

	path := fset.Arg(0)
	if path == "" {
		return errors.New("expected one argument: the .map to parse and graph")
	}

	qm, err := qmap.LoadFromFile(path)
	if err != nil {
		return fmt.Errorf("unable to read from map: %w", err)
	}

	GraphQMap(qm, os.Stdout)

	return nil
}
