// +build !novalidate,!nacl

package epubcheck

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pgaskin/epubtool/util"
)

//go:generate go run epubcheck_generate.go

// Run runs epubcheck on the specified file. Java is required to be in the PATH.
func Run(file string) (valid bool, err error) {
	td, err := ioutil.TempDir("", "epubcheck-*")
	if err != nil {
		return false, util.Wrap(err, "could not create temp dir for epubcheck")
	}
	defer os.RemoveAll(td)

	if err := exec.Command("java").Run(); err != nil && err.Error() != "exit status 1" {
		return false, util.Wrap(err, "error running java")
	}

	os.RemoveAll(td)
	if err := util.UnzipReader(bytes.NewReader(epubcheck), int64(len(epubcheck)), td); err != nil {
		return false, util.Wrap(err, "could not unpack epubcheck zip")
	}

	cmd := exec.Command("java", "-jar", filepath.Join(td, epubcheckJar), file)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = nil
	if err := cmd.Run(); err != nil {
		if err.Error() == "exit status 1" {
			return false, nil
		}
		return false, util.Wrap(err, "error running epubcheck")
	}

	return true, nil
}
