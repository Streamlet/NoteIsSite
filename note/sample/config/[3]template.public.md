# Template

```toml
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
```