package main

import (
	"errors"
	"strings"
)

type CLI struct {
	Args     []string
	Commands map[string]func() Command

	needHelp bool
	cmdArg   string
	paramArg string
}

var template *Template

func NewCLI() *CLI {
	return &CLI{
		Commands: map[string]func() Command{
			"generate": func() Command {
				return &GenerateCommand{
					Input: template,
				}
			},
			"validate": func() Command {
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

	if c.needHelp {
		c.printHelp()
		return nil
	}

	raw, ok := c.Commands[c.cmdArg]
	if !ok {
		c.printHelp()
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
		c.needHelp = true
		return errors.New("Too manay arguments specified")
	}
	if len(c.Args) < 2 {
		c.needHelp = true
		return errors.New("remotecmd requires atleast two arguments")
	}

	c.cmdArg = c.Args[0]
	c.paramArg = c.Args[1]

	_, ok := c.Commands[c.cmdArg]
	if ok != true {
		c.needHelp = true
		return errors.New("Command NOT supported, please see help")
	}

	if !strings.HasSuffix(c.paramArg, ".json") {
		c.needHelp = true
		return errors.New("template file should only be a JSON doc")
	}
	return nil

}

func (c *CLI) printHelp() {

}
