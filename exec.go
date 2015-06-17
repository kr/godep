package main

import (
	"log"
	"os"
	"os/exec"
)

var cmdExec = &Command{
	Usage: "exec command [args]",
	Short: "run a command with GOPATH set",
	Long: `
Exec sets the GOPATH to the vendored directory and then runs the specified command.
`,
	Run: runExec,
}

func runExec(cmd *Command, args []string) {
	if len(args) == 0 {
		log.Fatalln("Must specify command to execute")
	}

	c := exec.Command(args[0], args[1:]...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	c.Env = append(envNoGopath(), "GOPATH="+prepareGopath())

	err := c.Run()
	if err != nil {
		log.Fatal(err)
	}
}
