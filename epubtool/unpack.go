package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"

	et "github.com/geek1011/epubtool/epubtransform"
	"github.com/geek1011/epubtool/util"
)

func init() {
	commands = append(commands, &command{"unpack", "u", unpackMain})
}

func unpackMain(args []string, fs *pflag.FlagSet) int {
	// TODO: custom output dir
	help := fs.BoolP("help", "h", false, "Show this help text")
	fs.Parse(args)

	if *help || fs.NArg() != 2 {
		unpackHelp(args, fs)
		return 2
	}

	f := fs.Arg(1)
	fp, fe := util.SplitExt(f)
	if fe != ".epub" {
		fmt.Fprintf(os.Stderr, "Error: %s is not an epub file\n", f)
		return 1
	}

	fn := filepath.Base(fp)
	fmt.Printf("Unpacking %#v to %#v\n", f, fn)

	if err := et.New().Run(et.FileInput(f), et.DirOutput(fn)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func unpackHelp(args []string, fs *pflag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] epub_file\n\nOptions:\n", args[0])
	fs.PrintDefaults()
}
