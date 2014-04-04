package main

import (
	"fmt"
)

const (
	Version = "dev"
)

var cmdVersion = &Command{
	Run:   runVersion,
	Usage: "version",
	Short: "Display current version",
	Long: `
Display current version

Examples:

  godep version
`,
}

func init() {
}

func runVersion(cmd *Command, args []string) {
	fmt.Println("godep version:", Version)
}
