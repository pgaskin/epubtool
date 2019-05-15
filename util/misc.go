package util

import (
	"fmt"
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
