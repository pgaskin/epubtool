package util

import "fmt"

// Wrap wraps an error if not nil with a formatted string.
func Wrap(err error, format string, a ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %v", fmt.Sprintf(format, a...), err)
}
