package ssh

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"golang.org/x/crypto/ssh"
	"net"
	"quick-deploy/config"
)

// Client represents an SSH client
type Client struct {
	config *ssh.ClientConfig
	addr   string
	host   string
}

// getDefaultKeyFiles returns paths to common SSH private key files
func getDefaultKeyFiles() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return []string{
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_ecdsa"),
		filepath.Join(home, ".ssh", "id_dsa"),
	}
}

// readPrivateKey reads and parses an SSH private key
func readPrivateKey(keyPath string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key %s: %v", keyPath, err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key %s: %v", keyPath, err)
	}

	return ssh.PublicKeys(signer), nil
}

// getHostPort returns the host and port from the system's SSH config
func getHostPort(host string, configuredPort int) (string, int) {
	// Try running ssh with -G to get the resolved configuration
	cmd := exec.Command("ssh", "-G", host)
	output, err := cmd.Output()
	if err != nil {
		// If command fails, use the configured values
		return host, configuredPort
	}

	// Parse the output to find hostname and port
	lines := strings.Split(string(output), "\n")
	finalHost := host
	finalPort := configuredPort

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			switch parts[0] {
			case "hostname":
				finalHost = parts[1]
			case "port":
				if p, err := strconv.Atoi(parts[1]); err == nil {
					finalPort = p
				}
			}
		}
	}

	return finalHost, finalPort
}

// NewClient creates a new SSH client from server configuration
func NewClient(server *config.Server) (*Client, error) {
	var authMethods []ssh.AuthMethod

	// If password is provided, use password authentication
	if server.Password != "" {
		authMethods = append(authMethods, ssh.Password(server.Password))
	}

	// Try to use SSH keys
	for _, keyPath := range getDefaultKeyFiles() {
		if keyAuth, err := readPrivateKey(keyPath); err == nil {
			fmt.Printf("Successfully loaded SSH key: %s\n", keyPath)
			authMethods = append(authMethods, keyAuth)
		} else {
			fmt.Printf("Note: Could not load SSH key %s: %v\n", keyPath, err)
		}
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods available - please ensure SSH keys exist or provide password")
	}

	// Get the actual host and port from SSH config
	resolvedHost, resolvedPort := getHostPort(server.Host, server.Port)
	fmt.Printf("Resolved SSH config: %s:%d\n", resolvedHost, resolvedPort)

	config := &ssh.ClientConfig{
		User:            server.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", resolvedHost, resolvedPort)
	return &Client{
		config: config,
		addr:   addr,
		host:   server.Host, // Keep original host for logging
	}, nil
}

// ExecuteCommand executes a command on the remote server
func (c *Client) ExecuteCommand(command string, output io.Writer) error {
	fmt.Printf("Connecting to %s (%s)...\n", c.host, c.addr)

	// Create a custom dialer that respects SSH ProxyCommand
	dialer := net.Dialer{
		Timeout: c.config.Timeout,
	}

	// Try direct connection first
	conn, err := dialer.Dial("tcp", c.addr)
	if err != nil {
		// If direct connection fails, try using the system's ssh command as a proxy
		proxyCmd := exec.Command("ssh", "-W", c.addr, c.host)
		proxyConn, err := proxyCmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create proxy connection: %v", err)
		}
		proxyCmd.Stderr = os.Stderr
		if err := proxyCmd.Start(); err != nil {
			return fmt.Errorf("failed to start proxy command: %v", err)
		}
		conn = proxyConn.(net.Conn)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, c.addr, c.config)
	if err != nil {
		return fmt.Errorf("failed to establish SSH connection: %v", err)
	}

	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	fmt.Printf("Connected to %s, creating session...\n", c.host)
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	session.Stdout = output
	session.Stderr = output

	fmt.Printf("Executing command: %s\n", command)
	if err := session.Run(command); err != nil {
		return fmt.Errorf("failed to run command: %v", err)
	}

	return nil
}
