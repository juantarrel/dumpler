package main

import (
	"github.com/juantarrel/dumpler/cli"
	"os"
)

func main() {
	args := os.Args[1:]
	cli.Execute(args)
}
