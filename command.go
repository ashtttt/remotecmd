package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"log"

	"github.com/ashtttt/remotecmd/ssh"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/colorstring"
)

type Command interface {
	Run() error
	Set(string) error
}

type GenerateCommand struct {
	Input        *Template
	TemplateName string
}

type ValidateCommand struct {
	Input        *Template
	TemplateName string
}

type RunCommand struct {
	Input        *Template
	TemplateName string
}

func (g *GenerateCommand) Run() error {

	example := &Template{
		Bastion: &Bastion{
			Host: "bastion.bastion.com",
			User: "admin",
		},
		Remote: &Remote{
			Hosts: []string{"1.1.1.1", "0.0.0.0", "2.2.2.2"},
			Aws: &Aws{
				Nameprefix: "",
			},
			User: "ec2-user",
		},
		Commands: []string{"ls -lrth /usr/"},
	}

	templateJSON, _ := json.Marshal(example)
	data := []byte(string(templateJSON))

	filepath, err := os.Getwd()
	err = ioutil.WriteFile(filepath+"/"+g.TemplateName, data, 0644)

	if err != nil {
		return err
	}
	colorstring.Println("A sample template has been created in currect directory")
	return nil
}

func (g *GenerateCommand) Set(templateName string) error {
	g.TemplateName = templateName
	return nil
}

func (v *ValidateCommand) Run() error {

	filepath, err := os.Getwd()

	content, err := ioutil.ReadFile(filepath + "/" + v.TemplateName)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(string(content)), &v.Input)
	log.Println(v.Input.Commands)
	if err != nil {
		return err
	}
	return nil
}

func (v *ValidateCommand) Set(templateName string) error {
	v.TemplateName = templateName
	return nil
}

func (r *RunCommand) Run() error {
	filepath, err := os.Getwd()

	content, err := ioutil.ReadFile(filepath + "/" + r.TemplateName)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(string(content)), &r.Input)

	if err != nil {
		return err
	}

	if len(r.Input.Remote.Aws.Nameprefix) > 0 {
		data, err := getAWSNodes(r.Input.Remote.Aws.Nameprefix)
	}

	config := &ssh.Config{
		User:        r.Input.Remote.User,
		Nodes:       r.Input.Remote.Hosts,
		BastionHost: r.Input.Bastion.Host,
		BastionUser: r.Input.Bastion.User,
		Command:     "ls -lrth /usr/",
	}

	comm := ssh.New(config)

	_, err = comm.Run()

	if err != nil {
		return err
	}
	colorstring.Println("[green] Command completed in all hosts")
	return nil
}

func (r *RunCommand) Set(templateName string) error {
	r.TemplateName = templateName
	return nil
}

func getAWSNodes(namePrefix string) ([]string, error) {
	sess := session.Must(session.NewSession())
	svc := ec2.New(sess)

	params := &ec2.DescribeInstancesInput{
		DryRun: aws.Bool(true),
		Filters: []*ec2.Filter{
			{
				Name: "Name",
				Values: []*string{
					"devnx-idp",
				},
			},
		},
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return nil, err
	}
	fmt.Println(resp)

	return nil, nil

}
