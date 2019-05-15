package epubtransform

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/geek1011/epubtool/util"
)

// TODO: DirInput, DirOutput, AutoInput

// FileInput returns an InputFunc to read from an epub file.
func FileInput(file string) InputFunc {
	return func(epubdir string) error {
		zr, err := zip.OpenReader(file)
		if err != nil {
			return err
		}
		defer zr.Close()

		for _, f := range zr.File {
			if err := func(f *zip.File) error {
				path := filepath.Join(epubdir, f.Name)
				if f.FileInfo().IsDir() {
					os.MkdirAll(path, 0700)
				} else {
					os.MkdirAll(filepath.Dir(path), 0700)

					fr, err := f.Open()
					if err != nil {
						return err
					}
					defer fr.Close()

					f, err := os.Create(path)
					if err != nil {
						return err
					}
					defer f.Close()

					if _, err = io.Copy(f, fr); err != nil {
						return err
					}
				}
				return nil
			}(f); err != nil {
				return util.Wrap(err, "error extracting file %s", f.Name)
			}
		}

		return nil
	}
}

// DirOutput returns an OutputFunc write to a directory.
func DirOutput(dir string) OutputFunc {
	return func(epubdir string) error {
		if _, err := os.Stat(dir); err == nil {
			return fmt.Errorf("output directory %#v already exists", dir)
		}
		return util.CopyDir(epubdir, dir)
	}
}
