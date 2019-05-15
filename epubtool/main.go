package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/pflag"
)

var commands []*command

type command struct {
	Name  string
	Short string
	Main  func(args []string, fs *pflag.FlagSet) int
}

func main() {
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	cmdMap := map[string]*command{}
	cmdList := []string{}
	for _, cmd := range commands {
		for _, v := range []string{cmd.Name, cmd.Short} {
			if _, seen := cmdMap[v]; seen {
				panic("command already set: " + v)
			}
			cmdMap[v] = cmd
			cmdList = append(cmdList, v)
		}
	}

	if len(os.Args) < 2 {
		globalHelp(cmdList)
		os.Exit(2)
	}

	if cmd, ok := cmdMap[os.Args[1]]; !ok {
		globalHelp(cmdList)
		os.Exit(2)
	} else {
		args := append([]string{os.Args[0] + " " + os.Args[1]}, os.Args[2:]...)
		fs := pflag.NewFlagSet(args[0], pflag.ExitOnError)
		os.Exit(cmd.Main(args, fs))
	}
}

func globalHelp(cmdList []string) {
	fmt.Fprintf(os.Stderr, "Usage: %s (%s) [options] epub_path\n", os.Args[0], strings.Join(cmdList, "|"))
}
