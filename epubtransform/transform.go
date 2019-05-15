package epubtransform

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

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
	Desc string
	OPF  func(opf string) (string, error)
	Raw  func(epubdir string) error
	// TODO: OPFDoc, NCXDoc, NCX, ContentFile; only parse/encode doc if required
}

// New creates a new pipeline.
func New(transforms ...Transform) Pipeline {
	return Pipeline(transforms)
}

// Run runs the transform pipeline.
func (p Pipeline) Run(input InputFunc, output OutputFunc) error {
	epubdir, err := ioutil.TempDir("", "epub-*")
	if err != nil {
		return util.Wrap(err, "could not create temp dir")
	}
	defer os.RemoveAll(epubdir)

	if err := input(epubdir); err != nil {
		return util.Wrap(err, "could not run input")
	}

	if _, err := os.Stat(filepath.Join(epubdir, "META-INF", "container.xml")); err != nil {
		return errors.New("could not access META-INF/container.xml")
	}

	for _, transform := range p {
		if transform.OPF != nil {
			if err := transformOPF(epubdir, transform.OPF); err != nil {
				return util.Wrap(err, "could not run opf transform (%s)", transform.Desc)
			}
		}
		if transform.Raw != nil {
			if err := transform.Raw(epubdir); err != nil {
				return util.Wrap(err, "could not run raw transform (%s)", transform.Desc)
			}
		}
	}

	if output == nil {
		return nil
	}

	if err := output(epubdir); err != nil {
		return util.Wrap(err, "could not run output")
	}

	return nil
}

func transformOPF(epubdir string, fn func(opf string) (string, error)) error {
	op, err := getOPFPath(epubdir)
	if err != nil {
		return util.Wrap(err, "could not get opf path")
	}
	buf, err := ioutil.ReadFile(op)
	if err != nil {
		return util.Wrap(err, "could not read opf")
	}
	nopf, err := fn(string(buf))
	if err != nil {
		return err
	}
	nbuf := []byte(nopf)
	if !bytes.Equal(buf, nbuf) {
		if err := ioutil.WriteFile(op, nbuf, 0644); err != nil {
			return util.Wrap(err, "could not write new opf")
		}
	}
	return nil
}
