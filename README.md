# epubtool
[![Build Status](https://travis-ci.org/geek1011/epubtool.svg?branch=master)](https://travis-ci.org/geek1011/epubtool)

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
```