package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Servers     []Server    `yaml:"servers"`
	Deployments Deployments `yaml:"deployments"`
}

// Server represents a single server configuration
type Server struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password,omitempty"`
}

// Deployments represents the deployment configurations for all servers
type Deployments struct {
	Servers []DeploymentServer `yaml:"servers"`
}

// DeploymentServer represents deployment configuration for a specific server
type DeploymentServer struct {
	Name string   `yaml:"name"`
	Pipe Pipeline `yaml:"pipe"`
}

// Pipeline represents a sequence of commands to be executed
type Pipeline struct {
	LocalCommands  []Command `yaml:"local_commands"`
	RemoteCommands []Command `yaml:"remote_commands"`
}

// Command represents a single command to be executed
type Command struct {
	Command    string `yaml:"command"`
	WorkingDir string `yaml:"working_dir,omitempty"`
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
