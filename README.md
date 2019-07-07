# epubtool
A tool and library to manipulate ePub files. This rewrite is currently a work in progress.

## Examples

```sh
# Unpack an epub
$ epubtool u book.epub

# Pack an epub
$ epubtool p book.epub

# Add series metadata to an existing epub and format the OPF document
$ epubtool to --series "Series Name" --series-index 1 --beautify book.epub 

# Validate an epub using epubcheck
$ epubtool v book.epub

# Get the OPF document from an epub
$ epubtool d --opf book.epub

# You can also use an unpacked epub with the above commands
$ epubtool to --title "New Title" book-folder/

# Automatically rename all the books in a folder into another directory.
$ epubtool r --clean --output ./out/ --pattern "{{.title}} - {{.author}}.epub" *.epub
```

## Features
- Dump internal epub files (opf, ncx, etc).
- Pack/unpack epubs.
- Apply transformations to the OPF document.
- Validate an epub.
- Work with packed and unpacked epubs.
- Automatically rename epubs.
- Future:
  - Apply transformations on content files.
  - Get metadata from epubs.
  - Optimize CSS.
  - Automatic cleanup.
  - Generate new epubs from source files (markdown or html).
  - Diff two epubs.