// +build !novalidate,!nacl

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/pgaskin/epubtool/epubcheck"
)

func init() {
	commands = append(commands, &command{"validate", "v", "Run epubcheck on a book.", validateMain})
}

func validateMain(args []string, fs *pflag.FlagSet) int {
	// TODO: option to ignore certain errors
	help := fs.BoolP("help", "h", false, "Show this help text")
	fs.Parse(args)

	if *help || fs.NArg() != 2 {
		validateHelp(args, fs)
		return 2
	}

	fn := fs.Arg(1)
	if filepath.Ext(fn) != ".epub" {
		fmt.Fprintf(os.Stderr, "Error: %s is not an epub file\n", fn)
		return 1
	}

	fmt.Printf("Running epubcheck on %#v\n", fn)

	if valid, err := epubcheck.Run(fn); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	} else if !valid {
		fmt.Fprintf(os.Stderr, "Error: epub is not valid\n")
		return 1
	}
	return 0
}

func validateHelp(args []string, fs *pflag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] epub_file\n\nOptions:\n", args[0])
	fs.PrintDefaults()
}
