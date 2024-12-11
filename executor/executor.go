package executor

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"quick-deploy/config"
	"quick-deploy/ssh"
	"strings"
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

// shouldIgnoreError checks if the error from a command should be ignored
func shouldIgnoreError(cmd string, err error) bool {
	// 忽略kill命令的错误
	if strings.Contains(cmd, "kill") || strings.Contains(cmd, "pkill") {
		return true
	}
	return false
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

	err := command.Run()
	if err != nil && !shouldIgnoreError(cmd.Command, err) {
		return fmt.Errorf("local command failed: %v", err)
	}

	return nil
}

// ExecuteRemote executes a command on the remote server
func (e *Executor) ExecuteRemote(cmd *config.Command) error {
	fmt.Fprintf(e.output, "Executing remote command on %s: %s\n", e.server.Name, cmd.Command)

	err := e.ssh.ExecuteCommand(cmd.Command, e.output)
	if err != nil && !shouldIgnoreError(cmd.Command, err) {
		return fmt.Errorf("remote command failed: %v", err)
	}

	return nil
}

// ExecutePipeline executes a complete deployment pipeline
func (e *Executor) ExecutePipeline(pipeline *config.Pipeline) error {
	for _, cmd := range pipeline.Commands {
		var err error
		switch cmd.Type {
		case config.CommandTypeLocal:
			err = e.ExecuteLocal(&cmd)
		case config.CommandTypeRemote:
			err = e.ExecuteRemote(&cmd)
		default:
			return fmt.Errorf("unknown command type: %s", cmd.Type)
		}

		if err != nil {
			return fmt.Errorf("pipeline execution failed: %v", err)
		}
	}

	return nil
}
