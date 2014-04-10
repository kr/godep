package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var cmdUpdate = &Command{
	Usage: "update [packages]",
	Short: "use newer versions of selected packages",
	Long: `
Update the named dependency packages to the revision currently
installed in GOPATH. New code will be copied into Godeps and
the new revision id will be written to the manifest.

For more about specifying packages, see 'go help packages'.
`,
	Run: runUpdate,
}

func runUpdate(cmd *Command, args []string) {
	err := update(args)
	if err != nil {
		log.Fatalln(err)
	}
}

func update(args []string) error {
	if len(args) == 0 {
		args = []string{"."}
	}
	var g Godeps
	manifest := filepath.Join("Godeps", "Godeps.json")
	err := ReadGodeps(manifest, &g)
	if os.IsNotExist(err) {
		manifest = "Godeps"
		err = ReadGodeps(manifest, &g)
	}
	if err != nil {
		return err
	}
	for _, arg := range args {
		any := markMatches(arg, g.Deps)
		if !any {
			log.Println("not in manifest:", arg)
		}
	}
	deps, err := LoadVCSAndUpdate(g.Deps)
	if err != nil {
		return err
	}
	f, err := os.Create(manifest)
	if err != nil {
		return err
	}
	_, err = g.WriteTo(f)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	if manifest != "Godeps" {
		srcdir := filepath.FromSlash("Godeps/_workspace/src")
		copySrc(srcdir, deps)
	}
	return nil
}

// markMatches marks each entry in deps with an import path that
// matches pat. It returns whether any matches occurred.
func markMatches(pat string, deps []Dependency) (matched bool) {
	f := matchPattern(pat)
	for i, dep := range deps {
		if f(dep.ImportPath) {
			deps[i].matched = true
			matched = true
		}
	}
	return matched
}

// matchPattern(pattern)(name) reports whether
// name matches pattern.  Pattern is a limited glob
// pattern in which '...' means 'any string' and there
// is no other special syntax.
// Taken from $GOROOT/src/cmd/go/main.go.
func matchPattern(pattern string) func(name string) bool {
	re := regexp.QuoteMeta(pattern)
	re = strings.Replace(re, `\.\.\.`, `.*`, -1)
	// Special case: foo/... matches foo too.
	if strings.HasSuffix(re, `/.*`) {
		re = re[:len(re)-len(`/.*`)] + `(/.*)?`
	}
	reg := regexp.MustCompile(`^` + re + `$`)
	return func(name string) bool {
		return reg.MatchString(name)
	}
}

func LoadVCSAndUpdate(deps []Dependency) ([]Dependency, error) {
	var err1 error
	var paths []string
	for _, dep := range deps {
		if dep.matched {
			paths = append(paths, dep.ImportPath)
		}
	}
	if len(paths) == 0 {
		return nil, errors.New("no packages can be updated")
	}
	ps, err := LoadPackages(paths...)
	if err != nil {
		return nil, err
	}

	var tocopy []Dependency
	for _, pkg := range ps {
		if pkg.Error.Err != "" {
			log.Println(pkg.Error.Err)
			err1 = errors.New("error loading dependencies")
			continue
		}
		if pkg.Standard {
			continue
		}
		vcs, reporoot, err := VCSFromDir(pkg.Dir, filepath.Join(pkg.Root, "src"))
		if err != nil {
			log.Println(err)
			err1 = errors.New("error loading dependencies")
			continue
		}
		id, err := vcs.identify(pkg.Dir)
		if err != nil {
			log.Println(err)
			err1 = errors.New("error loading dependencies")
			continue
		}
		if vcs.isDirty(pkg.Dir, id) {
			log.Println("dirty working tree:", pkg.Dir)
			err1 = errors.New("error loading dependencies")
			continue
		}
		comment := vcs.describe(pkg.Dir, id)

		var dep *Dependency
		for i := range deps {
			if deps[i].ImportPath == pkg.ImportPath {
				dep = &deps[i]
			}
		}
		if dep == nil { // can't happen
			log.Println(pkg.ImportPath, "internal error")
			err1 = errors.New("error loading dependencies")
			continue
		}
		dep.Rev = id
		dep.Comment = comment
		dep.dir = pkg.Dir
		dep.ws = pkg.Root
		dep.root = filepath.ToSlash(reporoot)
		dep.vcs = vcs
		tocopy = append(tocopy, *dep)
	}
	if err1 != nil {
		return nil, err1
	}
	return tocopy, nil
}
