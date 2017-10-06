package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/beevik/etree"
)

// UnpackEPUB unpacks an epub
func UnpackEPUB(src, dest string, overwritedest bool) error {
	if src == "" {
		return errors.New("source must not be empty")
	}

	if dest == "" {
		return errors.New("destination must not be empty")
	}

	if !exists(src) {
		return fmt.Errorf(`source file "%s" does not exist`, src)
	}

	if exists(dest) {
		if !overwritedest {
			return fmt.Errorf(`destination "%s" already exists`, dest)
		}
		os.RemoveAll(dest)
	}

	src, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("error resolving absolute path of source: %s", err)
	}

	dest, err = filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("error resolving absolute path of destination: %s", err)
	}

	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, 0700)
		} else {
			os.MkdirAll(filepath.Dir(path), 0700)
			f, err := os.Create(path)
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

// PackEPUB packs an epub
func PackEPUB(src, dest string, overwritedest bool) error {
	if src == "" {
		return errors.New("source dir must not be empty")
	}

	if dest == "" {
		return errors.New("destination must not be empty")
	}

	if !exists(src) {
		return fmt.Errorf(`source dir "%s" does not exist`, src)
	}

	if exists(dest) {
		if !overwritedest {
			return fmt.Errorf(`destination "%s" already exists`, dest)
		}
		os.RemoveAll(dest)
	}

	src, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("error resolving absolute path of sourcedir: %s", err)
	}

	dest, err = filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("error resolving absolute path of destination: %s", err)
	}

	if !exists(filepath.Join(src, "META-INF", "container.xml")) {
		return fmt.Errorf("could not find META-INF/container.xml")
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("error creating destination file: %s", err)
	}
	defer f.Close()

	epub := zip.NewWriter(f)
	defer epub.Close()

	var addFile = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get the path of the file relative to the source folder
		relativePath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf(`error getting relative path of "%s"`, path)
		}

		if path == dest {
			// Skip if it is trying to pack itself
			return nil
		}

		if !info.Mode().IsRegular() {
			// Skip if not file
			return nil
		}

		if path == filepath.Join(src, "mimetype") {
			// Skip if it is the mimetype file
			return nil
		}

		// Create file in zip
		w, err := epub.Create(relativePath)
		if err != nil {
			return fmt.Errorf(`error creating file in epub "%s": %s`, relativePath, err)
		}

		// Open file on disk
		r, err := os.Open(path)
		if err != nil {
			return fmt.Errorf(`error reading file "%s": %s`, path, err)
		}
		defer r.Close()

		// Write file from disk to epub
		_, err = io.Copy(w, r)
		if err != nil {
			return fmt.Errorf(`error writing file to epub "%s": %s`, relativePath, err)
		}

		return nil
	}

	// Do not compress mimetype
	mimetypeWriter, err := epub.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	})

	if err != nil {
		return errors.New("error writing mimetype to epub")
	}

	_, err = mimetypeWriter.Write([]byte("application/epub+zip"))
	if err != nil {
		return errors.New("error writing mimetype to epub")
	}

	err = filepath.Walk(src, addFile)
	if err != nil {
		return fmt.Errorf("error adding file to epub: %s", err)
	}

	return nil
}

// EPUBMetadata reoresents the metadata of an epub book
type EPUBMetadata struct {
	Title       string `json:"title,omitempty"`
	Author      string `json:"author,omitempty"`
	Publisher   string `json:"publisher,omitempty"`
	Description string `json:"description,omitempty"`
	Series      struct {
		Name  string  `json:"name,omitempty"`
		Index float64 `json:"index,omitempty"`
	} `json:"series,omitempty"`
	// TODO: Add more fields
}

// GetEPUBMetadata gets the metadata of an ePub
func GetEPUBMetadata(src string) (EPUBMetadata, error) {
	dest, err := ioutil.TempDir("", "epubtool")
	if err != nil {
		return EPUBMetadata{}, errors.New("error creating temp dir")
	}
	defer os.RemoveAll(dest)

	md := &EPUBMetadata{}

	if err := UnpackEPUB(src, dest, true); err != nil {
		return *md, fmt.Errorf("error unpacking epub: %s", err)
	}

	rsk, err := os.Open(filepath.Join(dest, "META-INF", "container.xml"))
	if err != nil {
		return *md, fmt.Errorf("error parsing container.xml: %s", err)
	}
	defer rsk.Close()

	container := etree.NewDocument()
	_, err = container.ReadFrom(rsk)
	if err != nil {
		return *md, fmt.Errorf("error parsing container.xml: %s", err)
	}

	rootfile := ""
	for _, e := range container.FindElements("//rootfiles/rootfile[@full-path]") {
		rootfile = e.SelectAttrValue("full-path", "")
	}
	if rootfile == "" {
		return *md, fmt.Errorf("error parsing container.xml")
	}

	rrsk, err := os.Open(filepath.Join(dest, rootfile))
	if err != nil {
		return *md, fmt.Errorf("error parsing content.opf: %s", err)
	}
	defer rrsk.Close()

	opf := etree.NewDocument()
	_, err = opf.ReadFrom(rrsk)
	if err != nil {
		return *md, fmt.Errorf("error parsing content.opf: %s", err)
	}

	md.Title = filepath.Base(src)
	for _, e := range opf.FindElements("//title") {
		md.Title = e.Text()
		break
	}
	for _, e := range opf.FindElements("//creator") {
		md.Author = e.Text()
		break
	}
	for _, e := range opf.FindElements("//publisher") {
		md.Publisher = e.Text()
		break
	}
	for _, e := range opf.FindElements("//description") {
		md.Description = e.Text()
		break
	}
	for _, e := range opf.FindElements("//meta[@name='calibre:series']") {
		md.Series.Name = e.SelectAttrValue("content", "")
		break
	}
	for _, e := range opf.FindElements("//meta[@name='calibre:series_index']") {
		i, err := strconv.ParseFloat(e.SelectAttrValue("content", "0"), 64)
		if err == nil {
			md.Series.Index = i
			break
		}
	}

	return *md, nil
}
