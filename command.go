package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Commands lists the available commands and help topics.
// The order here is the order in which they are printed
// by 'godep help'.
var commands = []*Command{
	cmdSave,
	cmdGo,
	cmdGet,
	cmdPath,
	cmdRestore,
	cmdUpdate,
	cmdDiff,
}

// Command is an implementation of a godep command
// like godep save or godep go.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, args []string)

	// Usage is the one-line usage message.
	// The first word in the line is taken to be the command name.
	Usage string

	// Short is the short description shown in the 'godep help' output.
	Short string

	// Long is the long message shown in the
	// 'godep help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
}

// Name returns the name of a command.
func (c *Command) Name() string {
	name := c.Usage
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// UsageExit prints usage information and exits.
func (c *Command) UsageExit() {
	fmt.Fprintf(os.Stderr, "Usage: godep %s\n\n", c.Usage)
	fmt.Fprintf(os.Stderr, "Run 'godep help %s' for help.\n", c.Name())
	os.Exit(2)
}
