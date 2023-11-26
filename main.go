package main

import (
	"errors"
	"flag"
	"fmt"
	"goldutil/qmap"
	"os"
)

var help = `Usage: %s COMMAND [ARGS…]

Commands:
    entgraph MAP
        Outputs a graphviz digraph of entity caller/callee relationships from a
        .map file. ripent exports use the same format and can be read too.
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
	case "entgraph":
		return doEntGraph(args)
	}

	return nil
}

func doEntGraph(args []string) error {
	fset := flag.NewFlagSet("entgraph", flag.ContinueOnError)
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
