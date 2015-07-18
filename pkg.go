package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
)

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
	c := exec.Command("go", append(args, name...)...)
	r, err := c.StdoutPipe()
	if err != nil {
		return nil, errorWithCommand(err, c)
	}
	c.Stderr = os.Stderr
	err = c.Start()
	if err != nil {
		return nil, errorWithCommand(err, c)
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
	err = c.Wait()
	if err != nil {
		return nil, errorWithCommand(err, c)
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
