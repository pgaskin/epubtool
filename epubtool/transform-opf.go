package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	et "github.com/geek1011/epubtool/epubtransform"
)

func init() {
	commands = append(commands, &command{"transform-opf", "to", transformOPFMain})
}

func transformOPFMain(args []string, fs *pflag.FlagSet) int {
	// TODO: more metadata, generate uuid, remove metadata
	title := fs.StringP("title", "t", "", "Set dc:title")
	creator := fs.StringP("creator", "c", "", "Set dc:creator")
	description := fs.StringP("description", "d", "", "Set dc:description")
	series := fs.String("series", "", "Set calibre:series meta")
	seriesIndex := fs.Float64("series-index", 0, "Set calibre:series_index meta")
	dump := fs.Bool("dump", false, "Show OPF after transformations")
	dryRun := fs.Bool("dry-run", false, "Do not actually overwrite file")
	beautify := fs.Int("beautify", 4, "Indent the OPF by a number of spaces (0 to disable)")
	help := fs.BoolP("help", "h", false, "Show this help text")
	fs.Parse(args)

	if *help || fs.NArg() != 2 || !(*title != "" || *creator != "") {
		transformOPFHelp(args, fs)
		return 2
	}

	fn := fs.Arg(1)
	pipeline := et.New()

	if *title != "" {
		pipeline = append(pipeline, et.TransformTitle(*title))
	}
	if *creator != "" {
		pipeline = append(pipeline, et.TransformCreator(*creator))
	}
	if *description != "" {
		pipeline = append(pipeline, et.TransformDescription(*description))
	}
	if *series != "" {
		pipeline = append(pipeline, et.TransformOPFMetaElementContent(fmt.Sprintf("set series to %#v", *series), "calibre:series", *series))
	}
	if *seriesIndex > 0 {
		pipeline = append(pipeline, et.TransformOPFMetaElementContent(fmt.Sprintf("set series index to %v", *seriesIndex), "calibre:series_index", fmt.Sprint(*seriesIndex)))
	}
	if *beautify > 0 {
		pipeline = append(pipeline, et.TransformOPFBeautify(*beautify))
	}
	if *dump {
		pipeline = append(pipeline, et.Transform{
			Desc: "dump opf",
			OPF: func(opf string) (string, error) {
				fmt.Print(opf)
				return opf, nil
			},
		})
	}

	var out et.OutputFunc
	if !*dryRun {
		out = et.AutoOutput(fn)
	}

	if err := pipeline.Run(et.AutoInput(fn), out, true); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}

func transformOPFHelp(args []string, fs *pflag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] (epub_file|epub_dir)\n\nOptions:\n", args[0])
	fs.PrintDefaults()
}
