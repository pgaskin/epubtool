package main

//go:generate go-bindata data/...

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bclicn/color"
	"gopkg.in/urfave/cli.v1"
)

var version = "dev"

// Exists checks whether a path exists
func exists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

func pack(c *cli.Context) error {
	if len(c.Args()) > 2 {
		return errors.New("Too many arguments")
	}

	if len(c.Args()) < 1 {
		return errors.New("Not enough arguments")
	}

	src := c.Args().Get(0)
	dest := fmt.Sprintf("%s.epub", filepath.Base(src))

	if len(c.Args()) == 2 {
		dest = c.Args().Get(1)
	}

	if filepath.Ext(dest) != ".epub" {
		return errors.New("Outputfile extension must be .epub")
	}

	fmt.Printf("Source: %s\n", src)
	fmt.Printf("Destination: %s\n", dest)

	fmt.Printf("Packing ePub\n")

	err := PackEPUB(src, dest, true)
	if err != nil {
		return err
	}

	fmt.Printf(color.BLightGreen("Successfully packed ePub to \"%s\"\n"), dest)

	return nil
}
func unpack(c *cli.Context) error {
	if len(c.Args()) > 2 {
		return errors.New("Too many arguments")
	}

	if len(c.Args()) < 1 {
		return errors.New("Not enough arguments")
	}

	src := c.Args().Get(0)
	dest := strings.Replace(filepath.Base(src), ".epub", "", -1)

	if len(c.Args()) == 2 {
		dest = c.Args().Get(1)
	}

	if filepath.Ext(src) != ".epub" {
		return errors.New("Source file extension must be .epub")
	}

	fmt.Printf("Source: %s\n", src)
	fmt.Printf("Destination: %s\n", dest)

	fmt.Printf("Unpacking ePub\n")

	err := UnpackEPUB(src, dest, false)
	if err != nil {
		return err
	}

	fmt.Printf(color.BLightGreen("Successfully unpacked ePub to \"%s\"\n"), dest)

	return nil
}

func getMetadata(c *cli.Context) error {
	if len(c.Args()) > 1 {
		return errors.New("Too many arguments")
	}

	if len(c.Args()) < 1 {
		return errors.New("Not enough arguments")
	}

	src := c.Args().Get(0)

	if !exists(src) {
		return fmt.Errorf("File \"%s\" does not exist", src)
	}

	m, err := GetEPUBMetadata(src)
	if err != nil {
		return err
	}

	if m.Title != "" {
		fmt.Printf("       Title: %s\n", m.Title)
	}

	if m.Title != "" {
		fmt.Printf("      Author: %s\n", m.Author)
	}

	if m.Publisher != "" {
		fmt.Printf("   Publisher: %s\n", m.Publisher)
	}

	if m.Description != "" {
		fmt.Printf(" Description: %s\n", m.Description)
	}

	if m.Series.Name != "" {
		fmt.Printf("      Series: %s\n", m.Series.Name)
		fmt.Printf("Series Index: %v\n", m.Series.Index)
	}

	return nil
}

func Unzip(src, dest string) error {
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
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
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

func validate(c *cli.Context) error {
	if len(c.Args()) > 1 {
		return errors.New("Too many arguments")
	}

	if len(c.Args()) < 1 {
		return errors.New("Not enough arguments")
	}

	src := c.Args().Get(0)

	if !exists(src) {
		return fmt.Errorf("File \"%s\" does not exist", src)
	}

	td, err := ioutil.TempDir("", "epubtool")
	if err != nil {
		return fmt.Errorf("Could not create temp dir: %s", err)
	}
	defer os.RemoveAll(td)

	data, err := Asset("data/epubcheck.zip")
	if err != nil {
		return fmt.Errorf("Error extracting epubcheck: %s", err)
	}

	w, err := os.Create(filepath.Join(td, "epubcheckz"))
	if err != nil {
		return fmt.Errorf("Error extracting epubcheck: %s", err)
	}
	defer w.Close()

	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("Error extracting epubcheck: %s", err)
	}
	w.Close()

	err = Unzip(filepath.Join(td, "epubcheckz"), td)
	if err != nil {
		return fmt.Errorf("Error extracting epubcheck: %s", err)
	}

	err = exec.Command("java").Run()
	if err != nil {
		if fmt.Sprintf("%s", err) != "exit status 1" {
			return fmt.Errorf("Error running java: %s", err)
		}
	}

	cmd := exec.Command("java", "-jar", filepath.Join(td, "epubcheck.jar"), src)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = nil
	err = cmd.Run()
	if err != nil {
		if fmt.Sprintf("%s", err) == "exit status 1" {
			fmt.Println()
			return fmt.Errorf("ePub file is not valid")
		}
		return fmt.Errorf("Error running java: %s", err)
	}

	fmt.Printf("\nePub file is valid.\n")
	return nil
}

func main() {
	app := cli.NewApp()

	app.Name = "epubtool"
	app.Description = "A tool to manipulate ePub files"
	app.Version = version

	app.Commands = []cli.Command{
		{
			Name:        "pack",
			Aliases:     []string{"p"},
			Usage:       "Packs an ePub",
			ArgsUsage:   "SOURCEDIR [OUTPUTFILE]",
			Description: "Packs an ePub into a file. Will overwite OUTPUTFILE if it exists. SOURCEDIR should contain the unpacked epub. By default, the ePub will be packed into an epub file in the current directory with the name being the basename of the SOURCEDIR, but you can specify a custom OUTPUTFILE.",
			Action:      pack,
		},
		{
			Name:        "unpack",
			Aliases:     []string{"u"},
			Usage:       "Unpacks an ePub",
			ArgsUsage:   "INPUTFILE [OUTPUTDIR]",
			Description: "Unpacks an ePub into a directory. The directory must not already exist. By default, it will be unpacked into a new directory (with the basename of the INPUTFILE) in the current directory, but you can specify a custom OUTPUTDIR.",
			Action:      unpack,
		},
		{
			Name:        "get-metadata",
			Aliases:     []string{"gm"},
			Usage:       "Gets metadata of an ePub",
			ArgsUsage:   "INPUTFILE",
			Description: "Gets the metadata of an ePub and displays it.",
			Action:      getMetadata,
		},
		{
			Name:      "validate",
			Aliases:   []string{"v"},
			Usage:     "Validates an ePub with ePubCheck",
			ArgsUsage: "INPUTFILE",
			Action:    validate,
		},
	}

	app.Run(os.Args)
}
