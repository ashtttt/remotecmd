package ssh

import (
	"errors"
	"fmt"
	"net"
	"os"

	"time"

	"github.com/mitchellh/colorstring"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type comm struct {
	config *Config
	agent  agent.Agent
	client *ssh.Client
}

type Config struct {
	Nodes       []string
	User        string
	BastionHost string
	BastionUser string
	Command     string
}

func New(config *Config) *comm {
	return &comm{
		config: config,
	}
}

func (c *comm) Run() (int, error) {

	err := c.getBastion()
	if err != nil {
		return 1, err
	}
	chn := make(chan string, len(c.config.Nodes))
	erchn := make(chan error, len(c.config.Nodes))

	now := time.Now()
	timeout := time.After(1 * time.Minute)

	tempNodes := make([]string, len(c.config.Nodes))

	copy(tempNodes, c.config.Nodes)

	for _, node := range c.config.Nodes {
		go func(node string) {

			session, err := c.newSession(node)
			if err != nil {
				erchn <- err
			}
			session.Stdin = os.Stdin
			session.Stdout = os.Stdout
			session.Stderr = os.Stderr
			err = session.Run(c.config.Command)
			chn <- node
			defer c.client.Close()
			defer session.Close()
		}(node)
	}

	for i := 0; i < len(c.config.Nodes); {
		select {
		case node := <-chn:
			colorstring.Println("[yellow]completed execution on node " + node)
			tempNodes = removeItem(node, tempNodes)
			i++
		case error := <-erchn:
			return 1, error
		case <-time.After(10 * time.Second):
			fmt.Printf("still continuing after.... %d sec \n", time.Since(now)/time.Second)
		case <-timeout:
			colorstring.Println("[red]fallowing nodes did not get completed")
			for _, ip := range tempNodes {
				colorstring.Println("[red]" + ip)
			}
			return 1, errors.New("operation took more than a minute, exiting. There may be a un-fineshed nodes. Please check output")
		}
	}
	return 0, nil
}

func (c *comm) newSession(node string) (*ssh.Session, error) {

	agent.ForwardToAgent(c.client, c.agent)
	fmt.Printf("Connecting via bastion to host: %s \n", node)

	conn, err := c.client.Dial("tcp", node+":22")
	if err != nil {
		return nil, err
	}

	clientConfig := &ssh.ClientConfig{
		User: c.config.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(c.agent.Signers),
		},
	}

	sshConn, sshChan, req, err := ssh.NewClientConn(conn, node+":22", clientConfig)
	if err != nil {
		return nil, err
	}

	c.client = ssh.NewClient(sshConn, sshChan, req)

	session, err := c.client.NewSession()

	if err != nil {
		return nil, err
	}
	err = agent.RequestAgentForwarding(session)

	if err != nil {
		return nil, err
	}
	return session, nil

}

func (c *comm) getBastion() error {
	socketlocation := os.Getenv("SSH_AUTH_SOCK")
	if socketlocation == "" {
		return fmt.Errorf("%s", errors.New("ssh agent is NOT running. Please start to continue"))
	}

	agentConn, err := net.Dial("unix", socketlocation)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	c.agent = agent.NewClient(agentConn)
	if c.agent == nil {
		return fmt.Errorf("%s", errors.New("Could not create agent"))
	}

	bastionConfig := &ssh.ClientConfig{
		User: c.config.BastionUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(c.agent.Signers),
		},
	}

	fmt.Printf("Connecting to bastion: %s \n", c.config.BastionHost)

	c.client, err = ssh.Dial("tcp", c.config.BastionHost+":22", bastionConfig)
	if err != nil {
		return err

	}
	return nil
}

func removeItem(item string, arry []string) []string {
	for i, val := range arry {
		if val == item {
			arry = append(arry[:i], arry[i+1:]...)
		}
	}
	return arry
}
