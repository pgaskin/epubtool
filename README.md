# epubtool
[![Build Status](https://travis-ci.org/geek1011/epubtool.svg?branch=master)](https://travis-ci.org/geek1011/epubtool)

A tool to manipulate ePub files.

work-in-progress

## Features
- Pack
- Unpack
- Get metadata
- Coming soon:
    - Set metadata
    - Clean ePub
    - Add CSS
    - Add JS
    - Validate

## Usage
- `epubtool pack /path/to/unpacked/ebook/` will pack to `./ebook.epub`
- `epubtool pack /path/to/unpacked/ebook/ /place/to/put/somebook.epub` will pack to `/place/to/put/somebook.epub`
- `epubtool unpack /path/to/bookfile.epub` will unpack to `./bookfile/`
- `epubtool unpack /path/to/bookfile.epub /place/to/put/unpacked/` will unpack to `/place/to/put/unpacked/`
- `epubtool get-metadata /path/to/bookfile.epub` will display the metadata of `/path/to/bookfile.epub`