package epubtransform

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/beevik/etree"
	"github.com/geek1011/epubtool/util"
)

// Pipeline represents a series of transformations.
type Pipeline []Transform

// InputFunc puts an unpacked epub in epubdir (epubdir will always exist and be empty).
type InputFunc func(epubdir string) error

// OutputFunc writes the output epub from epubdir (epubdir will always exist).
type OutputFunc func(epubdir string) error

// Transform is a single transformation applied to an unpacked epub.
type Transform struct {
	Desc        string
	OPF         func(opf string) (newOPF string, err error)
	OPFDoc      func(opf *etree.Document) error
	Raw         func(epubdir string) error
	ContentFile func(relpath, html string) (newHTML string, err error)
	ContentDoc  func(relpath string, doc *goquery.Document) error // warning: don't use this with badly structured html (i.e. unclosed tags)
	// TODO: NCXDoc, NCX
}

// New creates a new pipeline.
func New(transforms ...Transform) Pipeline {
	return Pipeline(transforms)
}

// Run runs the transform pipeline.
func (p Pipeline) Run(input InputFunc, output OutputFunc, verbose bool) error {
	epubdir, err := ioutil.TempDir("", "epub-*")
	if err != nil {
		return util.Wrap(err, "could not create temp dir")
	}
	defer os.RemoveAll(epubdir)

	if verbose {
		fmt.Printf("Opening input\n")
	}
	if err := input(epubdir); err != nil {
		return util.Wrap(err, "could not run input")
	}

	if _, err := os.Stat(filepath.Join(epubdir, "META-INF", "container.xml")); err != nil {
		return errors.New("could not access META-INF/container.xml")
	}

	for i, transform := range p {
		if verbose {
			if transform.Desc != "" {
				fmt.Printf("Running transform %d: %s\n", i+1, transform.Desc)
			} else {
				fmt.Printf("Running transform %d\n", i+1)
			}
		}
		if transform.OPF != nil {
			if err := transformOPF(epubdir, transform.OPF); err != nil {
				return util.Wrap(err, "could not run opf transform (%s)", transform.Desc)
			}
		}
		if transform.OPFDoc != nil {
			if err := transformOPFDoc(epubdir, transform.OPFDoc); err != nil {
				return util.Wrap(err, "could not run opfdoc transform (%s)", transform.Desc)
			}
		}
		if transform.Raw != nil {
			if err := transform.Raw(epubdir); err != nil {
				return util.Wrap(err, "could not run raw transform (%s)", transform.Desc)
			}
		}
		if transform.ContentFile != nil {
			if err := transformContent(epubdir, transform.ContentFile); err != nil {
				return util.Wrap(err, "could not run content transform (%s)", transform.Desc)
			}
		}
		if transform.ContentDoc != nil {
			if err := transformContentDoc(epubdir, transform.ContentDoc); err != nil {
				return util.Wrap(err, "could not run contentdoc transform (%s)", transform.Desc)
			}
		}
	}

	if output == nil {
		if verbose {
			fmt.Printf("Skipping output\n")
		}
		return nil
	}

	if verbose {
		fmt.Printf("Writing output\n")
	}
	if err := output(epubdir); err != nil {
		return util.Wrap(err, "could not run output")
	}

	return nil
}

func transformFile(filename string, fn func(string) (string, error)) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return util.Wrap(err, "could not read file")
	}
	nstr, err := fn(string(buf))
	if err != nil {
		return err
	}
	if nbuf := []byte(nstr); !bytes.Equal(buf, nbuf) {
		if err := ioutil.WriteFile(filename, nbuf, 0644); err != nil {
			return util.Wrap(err, "could not write new file")
		}
	}
	return nil
}

func transformOPF(epubdir string, fn func(string) (string, error)) error {
	op, err := getOPFPath(epubdir)
	if err != nil {
		return util.Wrap(err, "could not get opf path")
	}
	return transformFile(op, fn)
}

func transformOPFDoc(epubdir string, fn func(*etree.Document) error) error {
	return transformOPF(epubdir, func(opf string) (string, error) {
		doc := etree.NewDocument()
		if err := doc.ReadFromString(opf); err != nil {
			return opf, err
		}
		if err := fn(doc); err != nil {
			return opf, err
		}
		nopf, err := doc.WriteToString()
		if err != nil {
			return opf, err
		}
		return nopf, nil
	})
}

func transformContent(epubdir string, fn func(string, string) (string, error)) error {
	files, err := util.MultiGlob(epubdir, "**/*.html", "**/*.xhtml", "**/*.htm")
	if err != nil {
		return err
	}
	for _, file := range files {
		relpath, err := filepath.Rel(epubdir, file)
		if err != nil {
			return err
		}
		if err := transformFile(file, func(str string) (string, error) {
			return fn(relpath, str)
		}); err != nil {
			return util.Wrap(err, "transform %#v", file)
		}
	}
	return nil
}

func transformContentDoc(epubdir string, fn func(string, *goquery.Document) error) error {
	return transformContent(epubdir, func(relpath string, str string) (string, error) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(str))
		if err != nil {
			return str, err
		}
		if err := fn(relpath, doc); err != nil {
			return str, err
		}
		nstr, err := doc.Html()
		if err != nil {
			return str, err
		}
		return nstr, nil
	})
}
