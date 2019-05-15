package epubtransform

import (
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
	Raw  func(epubdir string) error
	// TODO: OPFDoc, NCXDoc, OPF, NCX, ContentFile; only parse/encode doc if required
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
		if transform.Raw != nil {
			if err := transform.Raw(epubdir); err != nil {
				return util.Wrap(err, "could not run raw transform (%s)", transform.Desc)
			}
		}
	}

	if err := output(epubdir); err != nil {
		return util.Wrap(err, "could not run output")
	}

	return nil
}
