# goldutil
GoldSrc utilities.

```
Usage: goldutil COMMAND [ARGS…]

Commands:
    entity-graph MAP
        Outputs a graphviz digraph of entity caller/callee relationships from a
        .map file. ripent exports use the same format and can be read too.

    sprite-info SPR
        Prints parsed frame data from a sprite.

    sprite-extract [-dir DIR] SPR
        Outputs all frames of a sprite to the current directory.

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
```
