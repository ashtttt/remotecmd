package main

import (
	"os"

	"github.com/mitchellh/colorstring"
)

func main() {
	cli := NewCLI()
	cli.Args = os.Args[1:]

	err := cli.Run()

	if err != nil {
		colorstring.Println("[red]" + err.Error())
		if cli.NeedHelp == true {
			cli.PrintHelp()
		}
	}

	os.Exit(0)
}
