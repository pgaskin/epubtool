package main

import (
	"errors"
	"fmt"
	"os"
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
	}

	app.Run(os.Args)
}
