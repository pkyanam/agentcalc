package main

import (
	"github.com/pkyanam/agentcalc/internal/cli"
	"os"
)

var version = "dev"

func main() { os.Exit(cli.Main(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, version)) }
