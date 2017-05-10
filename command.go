package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

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
			Hosts: []string{},
			Aws: &Aws{
				Nameprefix: "",
			},
			User: "ec2-user",
		},
		Command: "ls -lrth /usr/",
	}

	templateJSON, _ := json.Marshal(example)
	data := []byte(string(templateJSON))

	filepath, err := os.Getwd()
	err = ioutil.WriteFile(filepath+"/"+g.TemplateName, data, 0644)

	if err != nil {
		return err
	}
	colorstring.Println("[green]a sample template has been placed in currect directory")
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
		instances, err := getAWSNodes(r.Input.Remote.Aws.Nameprefix)
		if err != nil {
			fmt.Println(err)
		}
		r.Input.Remote.Hosts = append(r.Input.Remote.Hosts, instances...)
	}

	config := &ssh.Config{
		User:        r.Input.Remote.User,
		Nodes:       r.Input.Remote.Hosts,
		BastionHost: r.Input.Bastion.Host,
		BastionUser: r.Input.Bastion.User,
		Command:     r.Input.Command,
	}

	comm := ssh.New(config)
	colorstring.Println("[yellow]command will be executed in fallowing nodes. Verify and confirm by typing yes")
	for _, ip := range r.Input.Remote.Hosts {
		colorstring.Printf("[yellow] %s \n", ip)
	}
	colorstring.Printf("Enter (yes) or (no):")
	var confirmation string
	fmt.Scanf("%s", &confirmation)

	if confirmation == "yes" {
		now := time.Now()
		_, err = comm.Run()

		if err != nil {
			return err
		}
		colorstring.Printf("[green]command executed on all nodes. Elapsed time: %d sec \n", time.Since(now)/time.Second)
	} else {
		colorstring.Println("[red]operation has been cancelled")
	}
	return nil
}

func (r *RunCommand) Set(templateName string) error {
	r.TemplateName = templateName
	return nil
}

func getAWSNodes(namePrefix string) ([]string, error) {
	var ips []string
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	svc := ec2.New(sess)

	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(namePrefix + "*"),
				},
			},
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("running"),
				},
			},
		},
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return nil, err
	}

	for idx := range resp.Reservations {
		for _, instance := range resp.Reservations[idx].Instances {

			ips = append(ips, *instance.PrivateIpAddress)
		}
	}
	return ips, nil

}
