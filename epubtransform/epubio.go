package epubtransform

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/geek1011/epubtool/util"
)

// TODO: AutoInput

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

// DirOutput returns an OutputFunc to write to a directory. The destination must not exist.
func DirOutput(dir string) OutputFunc {
	return func(epubdir string) error {
		if _, err := os.Stat(dir); err == nil {
			return fmt.Errorf("output directory %#v already exists", dir)
		}
		return util.CopyDir(epubdir, dir)
	}
}

// DirInput returns an InputFunc to read from an unpacked epub directory.
func DirInput(dir string) InputFunc {
	return func(epubdir string) error {
		os.RemoveAll(epubdir)
		return util.CopyDir(dir, epubdir)
	}
}

// FileOutput returns an OutputFunc write to a epub file. The destination must not exist.
func FileOutput(file string) OutputFunc {
	return func(epubdir string) error {
		f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return util.Wrap(err, "error creating destination file")
		}
		defer f.Close()

		zw := zip.NewWriter(f)
		defer zw.Close()

		if mimetypeWriter, err := zw.CreateHeader(&zip.FileHeader{
			Name:   "mimetype",
			Method: zip.Store, // Do not compress mimetype
		}); err != nil {
			return util.Wrap(err, "error writing mimetype to epub")
		} else if _, err = mimetypeWriter.Write([]byte("application/epub+zip")); err != nil {
			return util.Wrap(err, "error writing mimetype to epub")
		}

		if err := filepath.Walk(epubdir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(epubdir, path)
			if err != nil {
				return fmt.Errorf("error getting relative path of %#v", path)
			}

			// Skip if it is trying to pack itself, is not regular file, or is mimetype
			if path == epubdir || !info.Mode().IsRegular() || filepath.Base(path) == "mimetype" {
				return nil
			}

			fw, err := zw.Create(relPath)
			if err != nil {
				return util.Wrap(err, `error creating file %#v in epub`, relPath)
			}

			sf, err := os.Open(path)
			if err != nil {
				return util.Wrap(err, "error reading file %#v", path)
			}
			defer sf.Close()

			if _, err := io.Copy(fw, sf); err != nil {
				return util.Wrap(err, "error writing file %#v to epub", relPath)
			}

			return nil
		}); err != nil {
			return util.Wrap(err, "error creating epub")
		}

		return nil
	}
}
