# Note

```toml
# root directory for notes, can be relative to working directory, or absolute
note_root = "note/sample"

# category config file
# a directory will be treated as a note catogory only if there is a cate gory config file in it
category_config_file = "category.toml"

# pattern of notes filename, only matched files will be public
# the first capture group is the name of the file and will be the part of the url
note_file_pattern = "^(?:\\[.*?\\])*(.*)\\.public\\.(?:txt|html|md)$"
```