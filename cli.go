package main

import (
	"errors"
	"strings"

	"github.com/mitchellh/colorstring"
)

type CLI struct {
	Args     []string
	Commands map[string]func() Command

	NeedHelp bool
	cmdArg   string
	paramArg string
}

var template *Template

func NewCLI() *CLI {
	return &CLI{
		Commands: map[string]func() Command{
			"gen": func() Command {
				return &GenerateCommand{
					Input: template,
				}
			},
			"val": func() Command {
				return &ValidateCommand{
					Input: template,
				}
			},
			"run": func() Command {
				return &RunCommand{
					Input: template,
				}
			},
		},
	}
}

func (c *CLI) Run() error {

	err := c.validateArgs()
	if err != nil {
		return err
	}
	raw, ok := c.Commands[c.cmdArg]
	if !ok {
		c.PrintHelp()
	}
	command := raw()

	command.Set(c.paramArg)
	err = command.Run()

	if err != nil {
		return err
	}
	return nil
}

func (c *CLI) validateArgs() error {
	if len(c.Args) > 2 {
		c.NeedHelp = true
		return errors.New("Too manay arguments specified")
	}
	if len(c.Args) < 2 {
		c.NeedHelp = true
		return errors.New("remotecmd requires atleast two arguments")
	}

	c.cmdArg = c.Args[0]
	c.paramArg = c.Args[1]

	_, ok := c.Commands[c.cmdArg]
	if ok != true {
		c.NeedHelp = true
		return errors.New("Command NOT supported, please see help")
	}

	if !strings.HasSuffix(c.paramArg, ".json") {
		c.NeedHelp = true
		return errors.New("Template file should only be JSON doc")
	}
	return nil

}

func (c *CLI) PrintHelp() {
	var usage = `Usage: remotecmd [command...] <template.json>
Commands:
  -gen		Generates a sample template in working directory.
  -val		Validates the input template.
  -run		Runs the provided command on remote hosts.
`
	colorstring.Println("[red]" + usage)
}
