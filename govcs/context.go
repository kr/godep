// Copyright 2014 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package govcs

import ()

type Context struct {
	GOARCH        string   `json:",omitempty"` // target architecture
	GOOS          string   `json:",omitempty"` // target operating system
	GOROOT        string   `json:",omitempty"` // Go root
	GOPATH        string   `json:",omitempty"` // Go path
	CgoEnabled    bool     `json:",omitempty"` // whether cgo can be used
	UseAllFiles   bool     `json:",omitempty"` // use files regardless of +build lines, file names
	Compiler      string   `json:",omitempty"` // compiler to assume when computing target paths
	BuildTags     []string `json:",omitempty"` // build constraints to match in +build lines
	ReleaseTags   []string `json:",omitempty"` // releases the current release is compatible with
	InstallSuffix string   `json:",omitempty"` // suffix to use in the name of the install dir
}
