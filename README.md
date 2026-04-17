# goldutil
GoldSrc CLI utilities, released under the MIT license, see
[LICENSE.md](./LICENSE.md).

The documentation is available online at https://l-p.github.io/goldutil/ or by
running `$ goldutil help`.

Features:

- WAD and SPR to PNG.
- PNG to WAD and SPR.
- Entity call graph via Graphviz.
- .map file cleanup to automate TrenchBroom workflows.
- `materials.txt` remapping for BSPs with embedded textures.
- Entity pre-processor (see `map neat`).
- WAV looping.
- NOD extraction.
- Human-readable dumps of various formats.

```shell
$ goldutil -h
NAME:
   goldutil - GoldSrc modding utilities.

USAGE:
   goldutil [global options] [command [command options]]

VERSION:
   v1.5.0

DESCRIPTION:
   goldutil can read, modify, and write multiple file formats used by the GoldSrc (Half-Life) engine.
   See more detailed help with `goldutil CMD -h` or `goldutil CMD SUBCMD -h`.

COMMANDS:
   bsp      BSP (compiled maps) manipulation.
   fgd      Output the FGD to use with goldutil map neat.
   nod      NPC pathfinding nodes manipulation.
   map      Map pre-processing.
   mod      Misc modding utilities
   spr      Sprite manipulation.
   wad      Texture files manipulation.
   wav      Audio manipulation.
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```
