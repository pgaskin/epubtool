package util

import (
	"errors"
	"fmt"
	"testing"
)

func TestWrap(t *testing.T) {
	err := errors.New

	for _, c := range []struct {
		err      error
		format   string
		a        []interface{}
		expected error
	}{
		{nil, "testing", nil, nil},
		{err("wrapped"), "testing", nil, err("testing: wrapped")},
		{err("wrapped"), "testing %s", []interface{}{"formatted"}, err("testing formatted: wrapped")},
	} {
		wrapped := Wrap(c.err, c.format, c.a...)
		if (c.expected == nil && wrapped != nil) || fmt.Sprint(c.expected) != fmt.Sprint(wrapped) {
			t.Errorf("Wrap(%#v, %#v, %#v): expected '%v', got '%v'", c.err, c.format, c.a, c.expected, wrapped)
		}
	}
}

func TestSplitExt(t *testing.T) {
	for _, c := range [][]string{
		{"test", "test", ""},
		{"test.ext", "test", ".ext"},
		{"test.another.ext", "test.another", ".ext"},
		{"dir/test.another.ext", "dir/test.another", ".ext"},
		{"dir/test.another.ext/", "dir/test.another.ext/", ""},
	} {
		f, e := SplitExt(c[0])
		if f != c[1] {
			t.Errorf("%#v: expected file %#v, got %#v", c[0], c[1], f)
		}
		if e != c[2] {
			t.Errorf("%#v: expected ext %#v, got %#v", c[0], c[2], e)
		}
	}
}
