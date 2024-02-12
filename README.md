# goldutil
GoldSrc utilities.

```
Usage: goldutil COMMAND [ARGS…]

Commands:
    map-export [-cleanup-tb] MAP
        Exports a .map file the way TrenchBroom does, removing all layers
        marked as not exported.
        Output is written to stdout.

        Options:
            -cleanup-tb Removes properties added by TrenchBroom that are not
                        understood by the engine and spam the console with
                        errors.

    map-graph MAP
        Creates a graphviz digraph of entity caller/callee relationships from a
        .map file. ripent exports use the same format and can be read too.
        Output is written to stdout.

    sprite-info SPR
        Prints parsed frame data from a sprite.

    sprite-extract [-dir DIR] SPR
        Outputs all frames of a sprite to the current directory. The output
        files will be named after the original sprite file name plus a frame
        number suffix and an extension.

        Options:
            -dir DIR    Outputs frames to the specified directory instead of
                        the current one.

    sprite-create [-type TYPE] [-format FORMAT] FRAME0 [FRAMEX…]
        Creates a sprite from the given ordered list of PNG frames and writes
        it to the given SPR path.
        Input images must be 256 colors paletted PNGs. The palette of
        the first frame will be used, the other palettes are discarded and all
        frames will be interpreted using the first frame's palette.
        If the palette has under 256 colors it will be extended to 256,
        putting the last color of the palette in the 256th spot and remapping
        the image to match this updated palette. This matters for some texture
        formats.

        Options:
            -out SPR
                Path to the output .spr file.

            -type TYPE
                Sprite type, TYPE can be any one of:

                parallel           Always face camera. (Default)
                parallel-upright   Always face camera except for the locked Z axis.
                oriented           Orientation set by the level.
                parallel-oriented  Faces camera but can be rotated by the level.
                facing-upright     Like parallel upright but faces the player
                                   origin instead of the camera.

            -format FORMAT
                Texture format, determines how the palette is interpreted and the
                texture is rendered by the engine. FORMAT can be any one of:

                normal      256 colors sprite. (Default)
                additive    Additive 256 colors sprite.
                index-alpha Monochromatic sprite with 256 alpha levels, the base
                            color is determined by the last color on the palette.
                alpha-test  Transparent 255 colors sprite. The 256th color on the
                            palette will be rendered as fully transparent.

    wad-create -out WAD PATH [PATH…]
        Create a WAD file from a list of PNG files and directories. Directories
        are not scanned recursively and only PNG files are used.
        File base names (without extensions) are uppercased and used as texture
        names. This means that names exceeding 15 chars will trigger an error.

    wad-extract -out DIR WAD
        Extract a WAD file in the given directory as a bunch of PNG files.

    wad-info WAD
        Prints parsed data from a WAD file.
```
