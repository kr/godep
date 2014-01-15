package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Package struct {
	Dir        string
	Root       string
	ImportPath string
	Deps       []string
	Standard   bool

	TestImports []string

	Error struct {
		Err string
	}
}

// MustLoadPackages is like LoadPackages but it calls log.Fatal
// if an error occurs.
func MustLoadPackages(name ...string) []*Package {
	p, err := LoadPackages(name...)
	if err != nil {
		log.Fatal(err)
	}
	return p
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
		if info.Root == "" {
			goPath := filepath.SplitList(os.Getenv("GOPATH"))
			if len(goPath) > 0 {
				info.Root = goPath[0]
			}
		}
		a = append(a, info)
	}
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}
	return a, nil
}
