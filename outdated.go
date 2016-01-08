package main

import (
	"fmt"
	"log"
	"os"
)

var cmdOutdated = &Command{
	Name:  "outdated",
	Args:  "[-goversion] [packages]",
	Short: "List outdated packages, specify a package to see what has been commited.",
	Long: `
	List which packages are outdated.
	Specify a package to show the changes.

	For more about specifying packages, see 'go help packages'.
	`,
	Run: runOutdated,
}

func runOutdated(cmd *Command, args []string) {
	if len(args) > 0 {
		showOutdatedPackageDetails(args)
	} else {
		listOutdatedPackages()
	}
}

func listOutdatedPackages() {
	var paths []string
	var outdatedPackages []string
	gold, err := loadDefaultGodepsFile()
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalln(err)
		}
	}

	for _, dep := range gold.Deps {
		paths = append(paths, dep.ImportPath)
	}

	ps, err := LoadPackages(paths...)
	if err != nil {
		log.Fatalln(err)
	}

	for _, dep := range gold.Deps {
		for _, pkg := range ps {
			if dep.ImportPath == pkg.ImportPath {
				dep.pkg = pkg
				break
			}
		}
		vcs, err := VCSForImportPath(dep.ImportPath)
		if err != nil {
			log.Fatalln(err)
		}
		dirty := vcs.isDirty(dep.pkg.Dir, dep.Rev)
		if dirty {
			outdatedPackages = append(outdatedPackages, dep.ImportPath)
		}
	}
	for _, p := range outdatedPackages {
		fmt.Println(p)
	}
}

func showOutdatedPackageDetails(args []string) {
	gold, err := loadDefaultGodepsFile()
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatalln(err)
		}
	}

	ps, err := LoadPackages(args...)
	if err != nil {
		log.Fatalln(err)
	}

	for _, path := range args {
		var p *Package
		var d *Dependency

		for _, pkg := range ps {
			if path == pkg.ImportPath {
				p = pkg
				break
			}
		}

		for _, dep := range gold.Deps {
			if path == dep.ImportPath {
				d = &dep
				break
			}
		}
		vcs, err := VCSForImportPath(path)
		if err != nil {
			log.Fatalln(err)
		}

		out, err := vcs.runOutput(p.Dir, vcs.LogCmd, "rev", d.Rev)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("%s", out)
	}
}
