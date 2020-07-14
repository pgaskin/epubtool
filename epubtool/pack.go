package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"

	et "github.com/pgaskin/epubtool/epubtransform"
)

func init() {
	commands = append(commands, &command{"pack", "p", "Package a book.", packMain})
}

func packMain(args []string, fs *pflag.FlagSet) int {
	// TODO: custom output file
	help := fs.BoolP("help", "h", false, "Show this help text")
	fs.Parse(args)

	if *help || fs.NArg() != 2 {
		packHelp(args, fs)
		return 2
	}

	f := fs.Arg(1)
	of := strings.TrimRight(filepath.Clean(f), "/\\") + ".epub"

	fmt.Printf("Packing %#v to %#v\n", f, of)

	if err := et.New().Run(et.DirInput(f), et.FileOutput(of), false); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func packHelp(args []string, fs *pflag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] epub_dir\n\nOptions:\n", args[0])
	fs.PrintDefaults()
}
