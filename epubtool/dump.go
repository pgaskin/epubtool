package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	et "github.com/geek1011/epubtool/epubtransform"
)

func init() {
	commands = append(commands, &command{"dump", "d", "Show internal files from a book.", dumpMain})
}

func dumpMain(args []string, fs *pflag.FlagSet) int {
	// TODO: dump ncx, arbitrary file
	opf := fs.Bool("opf", false, "dump opf document")
	help := fs.BoolP("help", "h", false, "Show this help text")
	fs.Parse(args)

	if *help || fs.NArg() != 2 || !(*opf) {
		dumpHelp(args, fs)
		return 2
	}

	fn := fs.Arg(1)

	if err := et.New(et.Transform{
		Desc: "dump opf",
		OPF: func(opf string) (string, error) {
			fmt.Print(opf)
			return opf, nil
		},
	}).Run(et.AutoInput(fn), nil, false); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func dumpHelp(args []string, fs *pflag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] (epub_file|epub_dir)\n\nOptions:\n", args[0])
	fs.PrintDefaults()
}
