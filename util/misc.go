package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Wrap wraps an error if not nil with a formatted string.
func Wrap(err error, format string, a ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %v", fmt.Sprintf(format, a...), err)
}

// SplitExt splits a file path into a file and extension.
func SplitExt(path string) (string, string) {
	ext := filepath.Ext(path)
	return strings.TrimSuffix(path, ext), ext
}

// Unzip extracts a zip file to a directory, which must not exist.
func Unzip(src, dst string) error {
	f, err := os.OpenFile(src, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	return UnzipReader(f, fi.Size(), dst)
}

// UnzipReader unzips a zip file from a io.ReaderAt and a length to a directory, which must not exist.
func UnzipReader(r io.ReaderAt, l int64, dst string) error {
	zr, err := zip.NewReader(r, l)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("destination %#v must not exist", dst)
	}

	for _, f := range zr.File {
		if err := func(f *zip.File) error {
			op := filepath.Join(dst, f.Name)
			if f.FileInfo().IsDir() {
				os.MkdirAll(op, 0755)
			} else {
				os.MkdirAll(filepath.Dir(op), 0755)

				fr, err := f.Open()
				if err != nil {
					return err
				}
				defer fr.Close()

				f, err := os.Create(op)
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
			return Wrap(err, "error extracting file %s", f.Name)
		}
	}

	return nil
}