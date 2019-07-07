package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/beevik/etree"
	"github.com/geek1011/epubtool/epubtransform"
	"github.com/spf13/pflag"
)

func init() {
	commands = append(commands, &command{"rename", "r", "Rename a book.", renameMain})
}

func renameMain(args []string, fs *pflag.FlagSet) int {
	dryRun := fs.Bool("dry-run", false, "Dry run")
	overwrite := fs.Bool("overwrite", false, "Allow overwriting output files (DANGEROUS)")
	clean := fs.BoolP("clean", "c", false, "Replace [^A-Za-z0-9()_. -] with _")
	pattern := fs.StringP("pattern", "p", "{{.creator}} - {{.title}} {{if .series}}({{.series}} {{.series_index}}){{end}}.epub", "Pattern to rename files to")
	output := fs.StringP("output", "o", "", "Directory to copy output files to (must exist) (same as source if blank)")
	help := fs.BoolP("help", "h", false, "Show this help text")
	fs.Parse(args)

	if *help || fs.NArg() < 2 || *pattern == "" {
		renameHelp(args, fs)
		return 2
	}

	if strings.Contains(*pattern, "/") || strings.Contains(*pattern, "\\") {
		fmt.Fprintf(os.Stderr, "Error: directory creation is not supported yet.\n")
		return 2
	}

	fnRepl := strings.NewReplacer("/", "", "\\", "", "\"", "", "'", "")
	cleanRe := regexp.MustCompile(`((&#[0-9]+;)|([^A-Za-z0-9()_. -]))+`)
	tmpl, err := template.New("").Parse(*pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not parse pattern: %v.\n", err)
		return 2
	}

	fmt.Printf("Reading metadata:\n")
	meta := map[string]map[string]interface{}{}
	for _, fn := range fs.Args()[1:] {
		fmt.Printf("... %s\n", fn)
		if _, ok := meta[fn]; ok {
			fmt.Printf("    Warning: already processed, ignoring\n")
			continue
		}
		meta[fn] = map[string]interface{}{}
		if err := epubtransform.New(epubtransform.Transform{
			OPFDoc: func(opf *etree.Document) error {
				if el := opf.FindElement("//dc:title"); el != nil {
					meta[fn]["title"] = el.Text()
				}
				if el := opf.FindElement("//dc:creator"); el != nil {
					meta[fn]["creator"] = el.Text()
				}
				if el := opf.FindElement("//meta[@name='calibre:series']"); el != nil {
					meta[fn]["series"] = el.SelectAttrValue("content", "")
				}
				if el := opf.FindElement("//meta[@name='calibre:series_index']"); el != nil {
					meta[fn]["series_index"], _ = strconv.ParseFloat(el.SelectAttrValue("content", "0"), 64)
				}
				return nil
			},
		}).Run(epubtransform.FileInput(fn), nil, false); err != nil {
			fmt.Fprintf(os.Stderr, "    Error: error reading metadata: %v\n", err)
			return 1
		}
	}

	fmt.Printf("Generating filenames:\n")
	filenames := map[string]string{}
	seen := map[string]bool{}
	buf := bytes.NewBuffer(nil)
	for fn, m := range meta {
		fmt.Printf("... %#v\n", filepath.Base(fn))
		buf.Reset()
		if err := tmpl.Execute(buf, m); err != nil {
			fmt.Fprintf(os.Stderr, "    Error: error executing pattern: %v\n", err)
			return 1
		}

		filenames[fn] = buf.String()
		if *clean {
			filenames[fn] = cleanRe.ReplaceAllString(filenames[fn], "")
		}
		filenames[fn] = fnRepl.Replace(filenames[fn])
		if seen[filenames[fn]] {
			fmt.Fprintf(os.Stderr, "    Error: pattern results in duplicate filename\n")
			return 1
		}
		seen[filenames[fn]] = true

		fmt.Printf("    %#v\n", filenames[fn])
		if !strings.HasSuffix(filenames[fn], ".epub") {
			fmt.Printf("    Warning: generated filename does not end with .epub.\n")
		}
	}

	fmt.Printf("Renaming books:\n")
	for a, b := range filenames {
		var err error
		if a, err = filepath.Abs(a); err != nil {
			fmt.Fprintf(os.Stderr, "    Error: could not resolve absolute path to source: %v\n", err)
			return 1
		}

		if *output != "" {
			b = filepath.Join(*output, b)
			fmt.Printf("... Copy %#v -> %#v\n", a, b)
		} else {
			b = filepath.Join(filepath.Dir(a), b)
			fmt.Printf("... Move %#v -> %#v\n", a, b)
		}

		if b, err = filepath.Abs(b); err != nil {
			fmt.Fprintf(os.Stderr, "    Error: could not resolve absolute path to destination: %v\n", err)
			return 1
		}

		if a == b {
			fmt.Printf("    Warning: file already has desired name, skipping\n")
			continue
		}

		if !*overwrite {
			if _, err := os.Stat(b); !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "    Error: path %#v already exists (err=%v)\n", b, err)
				return 1
			}
		}

		if !*dryRun {
			in, err := os.OpenFile(a, os.O_RDONLY, 0)
			if err != nil {
				fmt.Fprintf(os.Stderr, "    Error: could not open file %#v: %v\n", a, err)
				return 1
			}
			out, err := os.OpenFile(b, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "    Error: could not create file %#v: %v\n", b, err)
				in.Close()
				return 1
			}

			if _, err := io.Copy(out, in); err != nil {
				fmt.Fprintf(os.Stderr, "    Error: could not write file: %v\n", err)
				out.Close()
				in.Close()
				os.Remove(b)
				return 1
			}

			if err := out.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "    Error: could not close output file: %v\n", err)
				in.Close()
				return 1
			}
			in.Close()
			if *output == "" {
				if err := os.Remove(a); err != nil {
					fmt.Fprintf(os.Stderr, "    Error: could not delete %#v: %v\n", a, err)
					return 1
				}
			}
		}
	}

	return 0
}

func renameHelp(args []string, fs *pflag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] epub_file...\n\nOptions:\n", args[0])
	fs.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Pattern:
  The provided pattern is applied using https://golang.org/pkg/text/template/. The
  following variables are provided:

  title          The content of dc:title
  creator        The content of dc:creator
  series         The content of meta[name=calibre:series]
  series_index   The content of meta[name=calibre:series_index]
`)
}
