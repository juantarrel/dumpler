package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/juantarrel/dumpler/cli"
	"os"
)

func main() {
	args := os.Args[1:]
	cli.Execute(args)
}
