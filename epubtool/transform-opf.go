package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	et "github.com/pgaskin/epubtool/epubtransform"
)

func init() {
	commands = append(commands, &command{"transform-opf", "to", "Modify a book's OPF metadata.", transformOPFMain})
}

func transformOPFMain(args []string, fs *pflag.FlagSet) int {
	// TODO: more metadata, generate uuid, remove metadata
	title := fs.StringP("title", "t", "", "Set dc:title")
	creator := fs.StringP("creator", "c", "", "Set dc:creator")
	description := fs.StringP("description", "d", "", "Set dc:description")
	publisher := fs.StringP("publisher", "p", "", "Set dc:publisher")
	series := fs.String("series", "", "Set calibre:series meta")
	seriesIndex := fs.Float64("series-index", 0, "Set calibre:series_index meta")
	dump := fs.Bool("dump", false, "Show OPF after transformations")
	meta := fs.StringToStringP("meta", "m", map[string]string{}, "Set one or more meta[name][content] tags (will remove if content is blank) (format name=content)")
	dryRun := fs.Bool("dry-run", false, "Do not actually overwrite file")
	beautify := fs.Int("beautify", 4, "Indent the OPF by a number of spaces (0 to disable)")
	help := fs.BoolP("help", "h", false, "Show this help text")
	fs.Parse(args)

	if *help || fs.NArg() != 2 || !(*title != "" || *creator != "" || *description != "" || *series != "" || *seriesIndex >= 0 || len(*meta) > 0 || *dump) {
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
	if *publisher != "" {
		pipeline = append(pipeline, et.TransformPublisher(*publisher))
	}
	for name, content := range *meta {
		txt := fmt.Sprintf("set meta[name=%#v][content=%#v]", name, content)
		if content == "" {
			txt = fmt.Sprintf("remove meta[name=%#v]", name)
		}
		pipeline = append(pipeline, et.TransformOPFMetaElementContent(txt, name, content))
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
