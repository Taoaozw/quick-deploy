package executor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"quick-deploy/config"
	"quick-deploy/ssh"
	"strings"
	"time"
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
	// 忽略rm命令的错误
	if strings.Contains(cmd, "rm -f") {
		return true
	}
	// 忽略systemctl stop的错误
	if strings.Contains(cmd, "systemctl stop") {
		return true
	}
	return false
}

// logCommand 记录命令执行的开始和结束
func (e *Executor) logCommand(cmdType string, cmd string) func(error) {
	startTime := time.Now()
	fmt.Fprintf(e.output, "==> [%s] Executing %s command: %s\n", 
		e.server.Name, cmdType, cmd)
	
	return func(err error) {
		duration := time.Since(startTime)
		if err != nil {
			if shouldIgnoreError(cmd, err) {
				fmt.Fprintf(e.output, "==> [%s] Command completed with ignored error (%.2fs)\n", 
					e.server.Name, duration.Seconds())
			} else {
				fmt.Fprintf(e.output, "==> [%s] Command failed (%.2fs): %v\n", 
					e.server.Name, duration.Seconds(), err)
			}
		} else {
			fmt.Fprintf(e.output, "==> [%s] Command completed successfully (%.2fs)\n", 
				e.server.Name, duration.Seconds())
		}
	}
}

// ExecuteLocal executes a local command
func (e *Executor) ExecuteLocal(cmd *config.Command) error {
	done := e.logCommand("local", cmd.Command)
	defer done(nil)

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
		done(err)
		return fmt.Errorf("local command failed: %v", err)
	}

	return nil
}

// ExecuteRemote executes a command on the remote server
func (e *Executor) ExecuteRemote(cmd *config.Command) error {
	done := e.logCommand("remote", cmd.Command)
	defer done(nil)

	err := e.ssh.ExecuteCommand(cmd.Command, e.output)
	if err != nil && !shouldIgnoreError(cmd.Command, err) {
		done(err)
		return fmt.Errorf("remote command failed: %v", err)
	}

	return nil
}

// ExecutePipeline executes a complete deployment pipeline
func (e *Executor) ExecutePipeline(pipeline *config.Pipeline) error {
	fmt.Fprintf(e.output, "\n==> Starting deployment on server: %s (%s)\n", 
		e.server.Name, e.server.Host)
	
	startTime := time.Now()
	
	for i, cmd := range pipeline.Commands {
		fmt.Fprintf(e.output, "\n==> Step %d/%d\n", i+1, len(pipeline.Commands))
		
		var err error
		switch cmd.Type {
		case config.CommandTypeLocal:
			err = e.ExecuteLocal(&cmd)
		case config.CommandTypeRemote:
			err = e.ExecuteRemote(&cmd)
		case config.CommandTypeScp:
			err = e.executeScpCommand(cmd.LocalPath, cmd.RemotePath)
		default:
			return fmt.Errorf("unknown command type: %s", cmd.Type)
		}

		if err != nil {
			fmt.Fprintf(e.output, "\n==> Deployment failed after %.2f seconds\n", 
				time.Since(startTime).Seconds())
			return fmt.Errorf("pipeline execution failed: %v", err)
		}
	}

	fmt.Fprintf(e.output, "\n==> Deployment completed successfully in %.2f seconds\n", 
		time.Since(startTime).Seconds())
	return nil
}

// executeScpCommand 执行 SCP 文件传输
func (e *Executor) executeScpCommand(localPath, remotePath string) error {
	fmt.Fprintf(e.output, "Copying file from %s to %s...\n", localPath, remotePath)
	
	// 检查本地文件是否存在
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer localFile.Close()

	// 获取文件信息
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// 确保远程路径存在
	remoteDir := filepath.Dir(remotePath)
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if err := e.ExecuteRemote(&config.Command{Command: mkdirCmd}); err != nil {
		return fmt.Errorf("failed to create remote directory: %v", err)
	}

	// 创建 SCP 会话
	session, err := e.ssh.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// 设置输出
	session.Stdout = e.output
	session.Stderr = e.output

	// 通过 SCP 传输文件
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		
		fmt.Fprintf(w, "C%#o %d %s\n", fileInfo.Mode().Perm(), fileInfo.Size(), filepath.Base(remotePath))
		io.Copy(w, localFile)
		fmt.Fprint(w, "\x00")
	}()

	if err := session.Run(fmt.Sprintf("scp -t %s", remotePath)); err != nil {
		return fmt.Errorf("scp failed: %v", err)
	}

	fmt.Fprintf(e.output, "Successfully copied file to %s\n", remotePath)
	return nil
}

// Close closes the executor and its SSH connection
func (e *Executor) Close() error {
	if e.ssh != nil {
		return e.ssh.Close()
	}
	return nil
}
