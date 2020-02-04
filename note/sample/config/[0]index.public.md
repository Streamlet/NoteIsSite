# Site Config

```toml
# Site Config

[server]

# port for the server to listening. If port is specified, sock MUST be empty string.
port = 80

# unix socket file for the server to listening. If sock is specified, port MUST be 0.
# sock = "/var/run/note_is_site.sock"

[template]

# root directory for html template, can be relative to working directory, or absolute
template_root = "template/sample"

# directories in template_root to be output staticly
static_dirs = [ "assets" ]

# files used by template system in template_root
index_template = "index.template.html"
category_template = "category.template.html"
content_template = "content.template.html"
404 = "404.html"
500 = "500.html"

[note]

# root directory for notes, can be relative to working directory, or absolute
note_root = "note/sample"

# category config file
# a directory will be treated as a note catogory only if there is a cate gory config file in it
category_config_file = "category.toml"

# pattern of notes filename, only matched files will be public
# the first capture group is the name of the file and will be the part of the url
note_file_pattern = "^(?:\\[.*?\\])*(.*)\\.public\\.(?:txt|html|md)$"

```

# Category Config

```toml
# Category Config

# name of the category
# if empty, defaults to the directory name in file system
name = "category_name"

# index file of the category
# the content of index file fills {{ .Content }} of the index template or category template
index = "index.md"
```