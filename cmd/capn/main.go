package main

import (
	"os"

	"github.com/iainlowe/capn/internal/cli"
)

func main() {
	c := cli.NewCLI()
	
	if err := c.Parse(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
