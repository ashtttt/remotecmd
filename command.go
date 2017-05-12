package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"errors"

	"strconv"

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
				Nameprefix: "test",
			},
			User: "ec2-user",
		},
		Command: "ls -l /usr/",
	}

	templateJSON, _ := json.MarshalIndent(example, "", "    ")
	data := []byte(string(templateJSON))

	filepath, err := os.Getwd()
	err = ioutil.WriteFile(filepath+"/"+g.TemplateName, data, 0644)

	if err != nil {
		return err
	}
	colorstring.Println("[green]Sample template has been placed in current directory")
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

	err = validateInput(v.Input)
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
	var awsAccountID string

	content, err := ioutil.ReadFile(filepath + "/" + r.TemplateName)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(string(content)), &r.Input)

	if err != nil {
		return err
	}

	err = validateInput(r.Input)
	if err != nil {
		return err
	}

	if len(r.Input.Remote.Aws.Nameprefix) > 0 {
		instances, awsid, err := getAWSNodes(r.Input.Remote.Aws.Nameprefix)
		awsAccountID = awsid
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
	colorstring.Printf("[yellow]Verify below details and confirm \n")
	colorstring.Printf("[yellow]Command: %s \n", r.Input.Command)
	if len(awsAccountID) > 0 {
		colorstring.Printf("[yellow]AWS Account: %s \n", awsAccountID)
	}

	colorstring.Printf("[yellow]Nodes:\n")
	for _, ip := range r.Input.Remote.Hosts {
		colorstring.Printf("[yellow] %s \n", ip)
	}
	colorstring.Printf("Enter (yes) to continue:")
	var confirmation string
	fmt.Scanf("%s", &confirmation)

	if confirmation == "yes" || confirmation == "y" {
		now := time.Now()
		_, err = comm.Run()

		if err != nil {
			return err
		}
		colorstring.Printf("[green]Command executed on all nodes. Elapsed time: %d sec \n", time.Since(now)/time.Second)
	} else {
		colorstring.Println("[red]Command execution cancelled!")
	}
	return nil
}

func (r *RunCommand) Set(templateName string) error {
	r.TemplateName = templateName
	return nil
}

func getAWSNodes(namePrefix string) ([]string, string, error) {
	os.Setenv("AWS_REGION", "us-east-1")
	var ips []string
	var accountID string
	sess := session.Must(session.NewSession())

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
		return nil, "", err
	}
	accountID = *resp.Reservations[0].OwnerId
	for idx := range resp.Reservations {
		for _, instance := range resp.Reservations[idx].Instances {
			ips = append(ips, *instance.PrivateIpAddress)
		}
	}
	return ips, accountID, nil

}

func validateInput(input *Template) error {
	valErrors := []string{}
	if len(input.Bastion.Host) <= 0 {
		valErrors = append(valErrors, "Bastion host is requred! Please verify input template \n")
	}
	if len(input.Bastion.User) <= 0 {
		valErrors = append(valErrors, "Bastion user name is required! Please verify input template \n")
	}
	if len(input.Remote.Hosts) <= 0 && len(input.Remote.Aws.Nameprefix) <= 0 {
		valErrors = append(valErrors, "Eithier remote hosts or AWS name-prefix required! Please verify input template \n")
	}
	if len(input.Remote.User) <= 0 {
		valErrors = append(valErrors, "Remote host user name is required! Please verify input template \n")
	}
	if len(input.Command) <= 0 {
		valErrors = append(valErrors, "Command is required! Please verify input template \n")
	}

	if len(valErrors) > 0 {
		var multiErrors string
		for i, err := range valErrors {
			multiErrors = multiErrors + strconv.Itoa(i) + "." + err
		}
		return errors.New(multiErrors)
	}
	return nil
}
