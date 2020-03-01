# Note

### How to hide some notes in note directory?
You could name the notes in different rules for those to publish and the ones not,
 and then use "note_file_pattern" option in [site_config](../config/site_config).
Only files matched the pattern will be listed and visited.

### How to to rewrite URL?
For note files, please use "note_file_pattern" option in [site_config](/config/site_config).
Place a capture in the regular expression,
and the first capture group will be the part of the url.

For directories, please use "name" option in [category_config](../config/category_config).

### How to sort notes and categories?
Notes and categories are in file system order by default.
You could rename files and directories with a numeric prefix (e.g. [0]first.md, [2]second.md, ...),
and hide the prefix by "note_file_pattern" option in [site_config](../config/site_config).

### How to use images for notes?
TODO

### How many file formats are supported for writing notes?
Currently only markdown is supported and recommended.

Files in other formats will be displayed as-is.
Thus, you could use .txt file for plain text, or write HTML contents in a .html file.

.doc/.docx are planned to be supported.

