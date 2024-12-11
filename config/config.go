package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// CommandType defines the type of command (local or remote)
type CommandType string

const (
	CommandTypeLocal  CommandType = "local"
	CommandTypeRemote CommandType = "remote"
)

// Command represents a command to be executed
type Command struct {
	Type       CommandType `yaml:"type"`
	Command    string      `yaml:"command"`
	WorkingDir string      `yaml:"working_dir,omitempty"`
}

// Pipeline represents a deployment pipeline
type Pipeline struct {
	Commands []Command `yaml:"commands"`
}

// Server represents a remote server configuration
type Server struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Deployment represents a deployment configuration for a server
type Deployment struct {
	Name string   `yaml:"name"`
	Pipe Pipeline `yaml:"pipe"`
}

// DeploymentConfig represents the deployment configuration for multiple servers
type DeploymentConfig struct {
	Servers []Deployment `yaml:"servers"`
}

// Config represents the complete configuration
type Config struct {
	Servers     []Server         `yaml:"servers"`
	Deployments DeploymentConfig `yaml:"deployments"`
}

// LoadConfig loads and parses the deploy.yaml file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return &cfg, nil
}

// validateConfig performs basic validation of the configuration
func validateConfig(cfg *Config) error {
	if len(cfg.Servers) == 0 {
		return fmt.Errorf("no servers defined in configuration")
	}

	// Create a map to check for duplicate server names
	serverNames := make(map[string]bool)
	for _, server := range cfg.Servers {
		if server.Name == "" {
			return fmt.Errorf("server name cannot be empty")
		}
		if serverNames[server.Name] {
			return fmt.Errorf("duplicate server name: %s", server.Name)
		}
		serverNames[server.Name] = true

		if server.Host == "" {
			return fmt.Errorf("host cannot be empty for server: %s", server.Name)
		}
		if server.Port <= 0 || server.Port > 65535 {
			return fmt.Errorf("invalid port number for server: %s", server.Name)
		}
		if server.Username == "" {
			return fmt.Errorf("username cannot be empty for server: %s", server.Name)
		}
	}

	return nil
}
