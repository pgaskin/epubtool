package epubtransform

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/beevik/etree"
	"github.com/pgaskin/epubtool/util"
)

func getOPFPath(epubdir string) (string, error) {
	cf, err := os.Open(filepath.Join(epubdir, "META-INF", "container.xml"))
	if err != nil {
		return "", util.Wrap(err, "error parsing container.xml")
	}
	defer cf.Close()

	container := etree.NewDocument()
	if _, err := container.ReadFrom(cf); err != nil {
		return "", util.Wrap(err, "error parsing container.xml")
	}

	var rootfile string
	for _, e := range container.FindElements("//rootfiles/rootfile[@full-path]") {
		rootfile = e.SelectAttrValue("full-path", "")
	}
	if rootfile == "" {
		return "", errors.New("error parsing container.xml: could not find rootfile full-path")
	}

	return filepath.Join(epubdir, rootfile), nil
}
