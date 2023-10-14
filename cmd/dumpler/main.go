package main

import (
	"github.com/juantarrel/dumper/cli"
	"os"
)

func main() {
	args := os.Args[1:]
	cli.Execute(args)
}
