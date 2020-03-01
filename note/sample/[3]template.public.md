# Template

### May I use template files with my custom file names?
Yes. Please refer to [site_config](../config/site_config), and notes for the following options:
```toml
index_template = "index.template.html"
category_template = "category.template.html"
content_template = "content.template.html"
404 = "404.html"
500 = "500.html"
```

### May I use another asset directory for static resources?
Yes. Please refer to [site_config](../config/site_config), and notes for the following options:
```toml
static_dirs = [ "assets" ]
```
You can put as many directories as you wish.

### How to reference data in template?
We use the golang template systems.
The original data structures are defined in template/template.go .
Samples are in template/sample.
