package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

var cmdForeach = &Command{
	Name:  "foreach",
	Short: "run a command in each dependency's path",
	Long: `
Foreach runs the provided command for each dependency path in GOPATH. This can
be useful for checking out the master branch after running "godep restore" to
avoid breaking "go get -u".

`,
	Run:          runForeach,
	OnlyInGOPATH: true,
}

func runForeach(cmd *Command, args []string) {
	g, err := loadDefaultGodepsFile()
	if err != nil {
		log.Fatalln(err)
	}

	if args == nil || len(args) < 1 {
		log.Fatalln("Must provide a command to foreach")
	}

	for _, dep := range g.Deps {
		c := exec.Command(args[0], args[1:]...)
		c.Dir = getRoot(dep.root, dep.ImportPath)

		fmt.Fprintf(os.Stderr, "\n%s:\n", dep.ImportPath)
		out, err := c.StdoutPipe()
		if err != nil {
			log.Fatalln(err)
		}

		serr, err := c.StderrPipe()
		if err != nil {
			log.Fatalln(err)
		}

		go io.Copy(os.Stdout, out)
		go io.Copy(os.Stderr, serr)

		err = c.Run()
		if err != nil {
			log.Fatalln(err)
		}
	}
}
