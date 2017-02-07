package main

import (
	"bufio"
	"io"
	"os"
	"strings"
)

var (
	appVendorFile = "appVendor"
)

func getAppPkgsFromReader(r io.Reader) ([]string, error) {
	s := bufio.NewScanner(r)
	var pkgs []string
	for s.Scan() {
		l := s.Text()
		l = strings.SplitN(l, "#", 2)[0]
		l = strings.SplitN(l, "//", 2)[0]
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		pkgs = append(pkgs, l)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return pkgs, nil
}

func getAppPkgs() ([]string, error) {
	f, err := os.Open(appVendorFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	return getAppPkgsFromReader(f)
}
