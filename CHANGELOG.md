# next
- Add command wav loop
- Add command bsp limits
- Build releases using Go v1.26
- Bump dependencies
- Allow sprite frames to have different sizes
- Fix missing color in index-alpha sprite extract
- Add --no-alpha flag to sprite extract
- Add --no-alpha flag to wad extract and make '{'-prefixed textures transparent by default

# v1.5.0
- Add "map neat" entity preprocessor command
- Add "fgd" command to use along with "map neat"
- Ensure lump names are lowercase
- Add `path_track` `message` property to targets when graphing entities
- Skip `multi_manager` `angle` property when graphing entities
- Update urfave/cli to v3
- Bring back default CLI help behavior
- Fix not handling property key with spaces in them
- Fix neat bailing when missing a titles.txt file

# v1.4.0
- Add command mod filter-wads

# v1.3.0
- Add command mod filter-materials

# v1.2.1
- Allow loading maps from stdin

# v1.2.0
- Add graph node extraction

# v1.1.1
- Replace help command with a single manpage

# v1.1.0
- BC break: migrated to urfave/cli, all commands are now sub-subcommands
- Add command bsp info
- Add command bsp remap-materials

# v1.0.1
- Fix nested groups not inheriting the `_tb_layer_omit_from_export` flag
- Add automated builds on Github

# v1.0.0
- WAD create/extract/info
- Sprite create/extract/info
- MAP processing for TrenchBroom
- MAP entities graph
