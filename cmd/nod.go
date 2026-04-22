package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/L-P/goldutil/goldsrc/nod"
	"github.com/L-P/goldutil/goldsrc/qmap"
)

func doNodExport(ctx context.Context, cmd *cli.Command) error {
	format, ok := map[string]nod.NodeFormat{
		"valve": nod.NodeFormatValve,
		"decay": nod.NodeFormatDecay,
	}[cmd.String("input-format")]
	if !ok {
		return errors.New("unrecognize .nod format")
	}

	f, err := os.Open(cmd.Args().Get(0))
	if err != nil {
		return fmt.Errorf("unable to open file for reading: %w", err)
	}
	defer f.Close() //nolint:errcheck // readonly

	nodes, links, err := nod.ReadNodes(f, format)
	if err != nil {
		return fmt.Errorf("unable to read nodes: %w", err)
	}

	original := cmd.Bool("original-positions")
	entities := make([]qmap.AnonymousEntity, 0, len(nodes)+len(links))
	for i, v := range nodes {
		entities = append(entities, qmap.AnonymousEntity{KVs: map[string]string{
			"classname":  v.ClassName(),
			"origin":     v.Position(original).String(),
			"targetname": fmt.Sprintf("node#%d", i),
		}})
	}

	for linkTypeBitID := range nod.LinkTypeBitMax {
		entities = append(entities, qmap.AnonymousEntity{KVs: map[string]string{
			"classname":            "func_group",
			"_tb_type":             "_tb_layer",
			"_tb_name":             fmt.Sprintf("hull#%d links (%s)", linkTypeBitID, nod.LinkTypeName(linkTypeBitID)),
			"_tb_id":               strconv.Itoa(linkTypeBitID + 1),
			"_tb_layer_sort_index": strconv.Itoa(linkTypeBitID + 1),
		}})

		for _, v := range links {
			if (v.LinkInfo & (1 << linkTypeBitID)) == 0 {
				continue
			}

			src := entities[v.SrcNode]
			src.KVs["target"] = fmt.Sprintf("node#%d", v.DstNode)
			src.KVs["_tb_layer"] = strconv.Itoa(linkTypeBitID + 1)
			entities = append(entities, src)
		}
	}

	out := qmap.New()
	if err := out.AddAnonymousEntities(entities...); err != nil {
		return fmt.Errorf("unable to append entities to output map: %w", err)
	}

	fmt.Fprintln(cmd.Writer, out.String())

	return nil
}
