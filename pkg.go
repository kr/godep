package main

import (
	"encoding/json"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Package represents a Go package.
type Package struct {
	Dir        string
	Root       string
	ImportPath string
	Deps       []string
	Standard   bool

	GoFiles        []string
	CgoFiles       []string
	IgnoredGoFiles []string

	TestGoFiles  []string
	TestImports  []string
	XTestGoFiles []string
	XTestImports []string

	Error struct {
		Err string
	}
}

// LoadPackages loads the named packages using go list -json.
// Unlike the go tool, an empty argument list is treated as
// an empty list; "." must be given explicitly if desired.
func LoadPackages(name ...string) (a []*Package, err error) {
	if len(name) == 0 {
		return nil, nil
	}
	args := []string{"list", "-e", "-json"}
	cmd := exec.Command("go", append(args, name...)...)
	r, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	d := json.NewDecoder(r)
	for {
		info := new(Package)
		err = d.Decode(info)
		if err == io.EOF {
			break
		}
		if err != nil {
			info.Error.Err = err.Error()
		}
		a = append(a, info)
	}
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}
	return a, nil
}

func walkExt(targetDir, ext string) (map[string]struct{}, error) {
	rmap := make(map[string]struct{})
	visit := func(path string, f os.FileInfo, err error) error {
		if f != nil {
			if !f.IsDir() {
				if filepath.Ext(path) == ext {
					if !filepath.HasPrefix(path, ".") && !strings.Contains(path, "/.") {
						wd, err := os.Getwd()
						if err != nil {
							return err
						}
						thepath := filepath.Join(wd, strings.Replace(path, wd, "", -1))
						rmap[thepath] = struct{}{}
					}
				}
			}
		}
		return nil
	}
	err := filepath.Walk(targetDir, visit)
	if err != nil {
		return nil, err
	}
	return rmap, nil
}

// https://github.com/golang/go/blob/master/src/go/build/syslist.go#L7
const goosList = "android darwin dragonfly freebsd linux nacl netbsd openbsd plan9 solaris windows "
const goarchList = "386 amd64 amd64p32 arm armbe arm64 arm64be ppc64 ppc64le mips mipsle mips64 mips64le mips64p32 mips64p32le ppc s390 s390x sparc sparc64 "
const appengineList = "appengine appenginevm"

func importDeps(dir string) (map[string]struct{}, error) {
	wm, err := walkExt(dir, ".go")
	if err != nil {
		return nil, err
	}
	fSize := len(wm)
	if fSize == 0 {
		return nil, nil
	}
	var mu sync.Mutex // guards the map
	fmap := make(map[string]struct{})
	done, errCh := make(chan struct{}), make(chan error)
	for fpath := range wm {
		go func(fpath string) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, fpath, nil, parser.ImportsOnly|parser.ParseComments)
			if err != nil {
				errCh <- err
				return
			}
			ignore := false
			for _, cc := range f.Comments {
				for _, v := range cc.List {
					if strings.HasPrefix(v.Text, "// +build ignore") {
						ignore = true
						break
					}
					if strings.HasPrefix(v.Text, "// +build") {
						p := strings.Replace(v.Text, "// +build ", "", -1)
						if !strings.Contains(goosList, p) && !strings.Contains(goarchList, p) && !strings.Contains(appengineList, p) {
							ignore = true
							break
						}
					}
				}
				if ignore {
					break
				}
			}
			if !ignore {
				for _, elem := range f.Imports {
					pv := strings.TrimSpace(strings.Replace(elem.Path.Value, `"`, "", -1))
					if pv == "C" || build.IsLocalImport(pv) || strings.HasPrefix(pv, ".") {
						continue
					}
					mu.Lock()
					fmap[pv] = struct{}{}
					mu.Unlock()
				}
			}
			done <- struct{}{}
		}(fpath)
	}
	i := 0
	for {
		select {
		case e := <-errCh:
			return nil, e
		case <-done:
			i++
			if i == fSize {
				close(done)
				return fmap, nil
			}
		}
	}
}

func LoadPackagesAll(name ...string) (a []*Package, err error) {
	a, err = LoadPackages(name...)
	if err != nil {
		return nil, err
	}
	for _, v := range a {
		// get dependencies from all Go files
		dm, err := importDeps(v.Dir)
		if err != nil {
			return nil, err
		}
		nDeps := make(map[string]struct{})
		for _, v := range v.Deps {
			nDeps[v] = struct{}{}
		}
		for k := range dm {
			nDeps[k] = struct{}{}
		}
		ds := []string{}
		for k := range nDeps {
			ds = append(ds, k)
		}
		sort.Strings(ds)
		v.Deps = ds
		v.GoFiles = append(v.GoFiles, v.IgnoredGoFiles...)
		v.IgnoredGoFiles = []string{}
	}
	return a, nil
}

func (p *Package) allGoFiles() (a []string) {
	a = append(a, p.GoFiles...)
	a = append(a, p.CgoFiles...)
	a = append(a, p.TestGoFiles...)
	a = append(a, p.XTestGoFiles...)
	a = append(a, p.IgnoredGoFiles...)
	return a
}
