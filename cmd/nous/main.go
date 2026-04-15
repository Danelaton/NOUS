package main

import (
	"fmt"
	"os"

	"github.com/nous-cli/nous/cmd/nous/cli"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("nous", version)
		return
	}
	cli.Execute()
}
