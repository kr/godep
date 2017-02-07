package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestAppPkgs(t *testing.T) {
	testAppVendorContents := `
// some slash comment
# some hash comment

github.com/tools/godep
github.com/golang/go // package with a slash comment
github.com/golang/tools # package with hash comment

`
	expectedPkgs := []string{
		"github.com/tools/godep",
		"github.com/golang/go",
		"github.com/golang/tools",
	}
	r := strings.NewReader(testAppVendorContents)
	pkgs, err := getAppPkgsFromReader(r)
	if err != nil {
		t.Errorf("Failed to parse %q: %v", testAppVendorContents, err)
	}
	if !reflect.DeepEqual(pkgs, expectedPkgs) {
		t.Errorf("App pkgs = %v want %v", pkgs, expectedPkgs)
	}
}
