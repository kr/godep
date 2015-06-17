package main

import (
	"fmt"
	"log"
	"os"
)

var cmdOutdated = &Command{
	Usage: "outdated",
	Short: "check to see if any dependencies are different",
	Long: `
Outdated compares the Godeps-specified version of each package against the versions in GOPATH.
`,
	Run: runOutdated,
}

func runOutdated(cmd *Command, args []string) {
	g, err := ReadAndLoadGodeps(findGodepsJSON())
	if err != nil {
		log.Fatalln(err)
	}
	hadError := false
	for _, dep := range g.Deps {
		err := compare(dep)
		if err != nil {
			log.Println("compare:", err)
			hadError = true
		}
	}
	if hadError {
		os.Exit(1)
	}
}

// compare looks at the specified dependency and compares it to the
// current version.
func compare(dep Dependency) error {
	ps, err := LoadPackages(dep.ImportPath)
	if err != nil {
		return err
	}
	pkg := ps[0]
	if !dep.vcs.exists(pkg.Dir, dep.Rev) {
		return fmt.Errorf("%s: required revision %s doesn't exist", dep.ImportPath, dep.Rev)
	}
	id, err := dep.vcs.identify(pkg.Dir)
	if err != nil {
		return err
	}
	if id != dep.Rev {
		return fmt.Errorf("%s: required revision %s is not checked out", dep.ImportPath, dep.Rev)
	}
	return nil
}
