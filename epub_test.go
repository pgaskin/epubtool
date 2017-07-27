package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackUnpack(t *testing.T) {
	td, err := ioutil.TempDir("", "epubtool")
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("%v", err), "Could not create temp dir")
	}
	defer os.RemoveAll(td)

	assert.Nil(t, PackEPUB(filepath.Join("testdata", "books", "test1"), filepath.Join(td, "test1.epub"), true), "packepub should not return an error")
	assert.True(t, exists(filepath.Join(td, "test1.epub")), "output epub should exist")

	assert.Nil(t, UnpackEPUB(filepath.Join(td, "test1.epub"), filepath.Join(td, "test1"), true), "unpackepub should not return an error")
	assert.True(t, exists(filepath.Join(td, "test1")), "output dir should exist")
	assert.True(t, exists(filepath.Join(td, "test1", "META-INF", "container.xml")), "META-INF/container.xml should exist in output dir")
}

func TestGetEPUBMetadata(t *testing.T) {
	// TODO
}
