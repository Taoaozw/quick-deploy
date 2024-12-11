package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"bufio"
	"quick-deploy/config"
	"golang.org/x/crypto/ssh"
)

// SSHConfig represents a parsed SSH config entry
type SSHConfig struct {
	Host         string
	HostName     string
	User         string
	IdentityFile string
}

// parseSSHConfig parses the SSH config file
func parseSSHConfig(configPath string) (map[string]*SSHConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	configs := make(map[string]*SSHConfig)
	var currentConfig *SSHConfig

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		value := parts[1]

		switch key {
		case "host":
			currentConfig = &SSHConfig{Host: value}
			configs[value] = currentConfig
		case "hostname":
			if currentConfig != nil {
				currentConfig.HostName = value
			}
		case "user":
			if currentConfig != nil {
				currentConfig.User = value
			}
		case "identityfile":
			if currentConfig != nil {
				// 展开 ~ 到用户主目录
				if strings.HasPrefix(value, "~/") {
					homeDir, err := os.UserHomeDir()
					if err == nil {
						value = filepath.Join(homeDir, value[2:])
					}
				}
				currentConfig.IdentityFile = value
			}
		}
	}

	return configs, scanner.Err()
}

// Client represents an SSH client
type Client struct {
	client *ssh.Client
	config *config.Server
}

// NewClient creates a new SSH client
func NewClient(server *config.Server) (*Client, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	// 读取 SSH 配置文件
	sshConfigs, err := parseSSHConfig(filepath.Join(homeDir, ".ssh", "config"))
	if err != nil {
		// 如果读取配置文件失败，记录错误但继续执行
		fmt.Fprintf(os.Stderr, "Warning: failed to read SSH config: %v\n", err)
	}

	var authMethods []ssh.AuthMethod
	var targetHost string = server.Host
	
	// 检查是否有匹配的 SSH 配置
	if sshConfig, ok := sshConfigs[server.Host]; ok {
		// 使用配置文件中的 HostName
		if sshConfig.HostName != "" {
			targetHost = sshConfig.HostName
		}
		
		// 使用配置的身份文件
		if sshConfig.IdentityFile != "" {
			if key, err := os.ReadFile(sshConfig.IdentityFile); err == nil {
				if signer, err := ssh.ParsePrivateKey(key); err == nil {
					authMethods = append(authMethods, ssh.PublicKeys(signer))
				}
			}
		}
	}

	// 如果没有找到配置或配置的密钥无效，尝试默认密钥
	if len(authMethods) == 0 {
		defaultKeyPaths := []string{
			filepath.Join(homeDir, ".ssh", "id_rsa"),
			filepath.Join(homeDir, ".ssh", "id_ed25519"),
		}

		for _, keyPath := range defaultKeyPaths {
			if key, err := os.ReadFile(keyPath); err == nil {
				if signer, err := ssh.ParsePrivateKey(key); err == nil {
					authMethods = append(authMethods, ssh.PublicKeys(signer))
				}
			}
		}
	}

	// 如果配置了密码，也添加密码认证
	if server.Password != "" {
		authMethods = append(authMethods, ssh.Password(server.Password))
	}

	// 如果没有找到任何认证方法，返回错误
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods available")
	}

	sshConfig := &ssh.ClientConfig{
		User: server.Username,
		Auth: authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", targetHost, server.Port), sshConfig)
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

	return session.Run(command)
}

// NewSession creates a new SSH session
func (c *Client) NewSession() (*ssh.Session, error) {
	return c.client.NewSession()
}

// Close closes the SSH connection
func (c *Client) Close() error {
	return c.client.Close()
}
