package ssh

import (
	"fmt"
	"io"
	"quick-deploy/config"
	"golang.org/x/crypto/ssh"
)

// Client represents an SSH client
type Client struct {
	client *ssh.Client
	config *config.Server
}

// NewClient creates a new SSH client
func NewClient(server *config.Server) (*Client, error) {
	sshConfig := &ssh.ClientConfig{
		User: server.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(server.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", server.Host, server.Port), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %v", err)
	}

	return &Client{
		client: client,
		config: server,
	}, nil
}

// ExecuteCommand executes a command on the remote server
func (c *Client) ExecuteCommand(command string, output io.Writer) error {
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	session.Stdout = output
	session.Stderr = output

	if err := session.Run(command); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	return nil
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
