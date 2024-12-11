package executor

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"quick-deploy/config"
	"quick-deploy/ssh"
)

// Executor handles command execution both locally and remotely
type Executor struct {
	server *config.Server
	ssh    *ssh.Client
	output io.Writer
}

// NewExecutor creates a new executor for a server
func NewExecutor(server *config.Server, output io.Writer) (*Executor, error) {
	sshClient, err := ssh.NewClient(server)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client: %v", err)
	}

	return &Executor{
		server: server,
		ssh:    sshClient,
		output: output,
	}, nil
}

// ExecuteLocal executes a local command
func (e *Executor) ExecuteLocal(cmd *config.Command) error {
	fmt.Fprintf(e.output, "Executing local command: %s\n", cmd.Command)

	command := exec.Command("sh", "-c", cmd.Command)
	if cmd.WorkingDir != "" {
		absPath, err := filepath.Abs(cmd.WorkingDir)
		if err != nil {
			return fmt.Errorf("failed to resolve working directory: %v", err)
		}
		command.Dir = absPath
	}

	command.Stdout = e.output
	command.Stderr = e.output

	if err := command.Run(); err != nil {
		return fmt.Errorf("local command failed: %v", err)
	}

	return nil
}

// ExecuteRemote executes a command on the remote server
func (e *Executor) ExecuteRemote(cmd *config.Command) error {
	fmt.Fprintf(e.output, "Executing remote command on %s: %s\n", e.server.Name, cmd.Command)

	if err := e.ssh.ExecuteCommand(cmd.Command, e.output); err != nil {
		return fmt.Errorf("remote command failed: %v", err)
	}

	return nil
}

// ExecutePipeline executes a complete deployment pipeline
func (e *Executor) ExecutePipeline(pipeline *config.Pipeline) error {
	// Execute local commands
	for _, cmd := range pipeline.LocalCommands {
		if err := e.ExecuteLocal(&cmd); err != nil {
			return err
		}
	}

	// Execute remote commands
	for _, cmd := range pipeline.RemoteCommands {
		if err := e.ExecuteRemote(&cmd); err != nil {
			return err
		}
	}

	return nil
}
