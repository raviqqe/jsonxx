package main

import (
	"fmt"
	"os"

	"github.com/raviqqe/jsonxx/command"
)

func main() {
	if err := command.Command(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}