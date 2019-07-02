package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mattn/go-zglob"
)

// Copy copies a file or directory to a nonexistent destination, preserving the
// permissions of its contents.
func Copy(src, dst string) error {
	if fi, err := os.Stat(src); err != nil {
		return err
	} else if fi.IsDir() {
		return CopyDir(src, dst)
	} else {
		return CopyFile(src, dst)
	}
}

// CopyDir copies a directory tree to a nonexistent destination, preserving the
// permissions of its contents. The destination must not exist. Symlinks are not
// supported.
func CopyDir(src, dst string) error {
	if fi, err := os.Stat(src); err != nil {
		return Wrap(err, "error reading source")
	} else if !fi.IsDir() {
		return fmt.Errorf("not a dir: %#v", src)
	}

	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("destination %#v already exists", dst)
	}

	if err := os.Mkdir(dst, os.ModePerm); err != nil {
		return Wrap(err, "could not create dest dir")
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return Wrap(err, "could not read dir %#v", src)
	}

	for _, entry := range entries {
		esrcp, edstp := filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name())
		// err is returned as-is to prevent really wrong messages in failures in deep structures
		if entry.IsDir() {
			if err = CopyDir(esrcp, edstp); err != nil {
				return err
			}
		} else {
			if err := CopyFile(esrcp, edstp); err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile copies a file and its permissions. The destination must not exist.
func CopyFile(src, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}

	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_EXCL, fi.Mode())
	if err != nil {
		return err
	}
	defer df.Close()

	if _, err = io.Copy(df, sf); err != nil {
		return err
	}

	return nil
}

// MultiGlob globs for multiple patterns.
func MultiGlob(base string, patterns ...string) ([]string, error) {
	var res []string
	for _, pattern := range patterns {
		v, err := zglob.Glob(filepath.Join(base, pattern))
		if err != nil {
			return res, err
		}
		res = append(res, v...)
	}
	return res, nil
}
