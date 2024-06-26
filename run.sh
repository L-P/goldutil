#!/usr/bin/env sh

dir="/home/anon/.dotfiles/.local/share/Steam/steamapps/common/Half-Life"
make && goldutil bsp remap-materials \
    --original-materials "$dir/valve/sound/materials.txt" \
    --replacement-materials "$dir/valve_addon/sound/makkon.materials.txt" \
    --out "$dir/valve_addon/maps/dm_solvent.remapped.bsp" \
    "$dir/valve_addon/maps/dm_solvent.bsp"
