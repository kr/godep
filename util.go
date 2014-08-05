package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Returns true if path definitely exists; false if path doesn't
// exist or is unknown because of an error.
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Runs a command in dir.
// The name and args are as in exec.Command.
// Stdout, stderr, and the environment are inherited
// from the current process.
func runIn(dir, name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	return errorWithCommand(err, c)
}

// command is like exec.Command, but the returned
// Cmd inherits stderr from the current process, and
// elements of args may be either string or []string.
func command(name string, args ...interface{}) *exec.Cmd {
	var a []string
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			a = append(a, v)
		case []string:
			a = append(a, v...)
		}
	}
	c := exec.Command(name, a...)
	c.Stderr = os.Stderr
	return c
}

func errorWithCommand(err error, cmd *exec.Cmd) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("Error running `%s`: %s", strings.Join(cmd.Args, " "), err.Error())
}
