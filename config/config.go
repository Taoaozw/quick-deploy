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
	CommandTypeScp    CommandType = "scp"
)

// Command represents a command to be executed
type Command struct {
	Type       CommandType `yaml:"type"`
	Command    string      `yaml:"command,omitempty"`
	WorkingDir string      `yaml:"working_dir,omitempty"`
	LocalPath  string      `yaml:"local_path,omitempty"`
	RemotePath string      `yaml:"remote_path,omitempty"`
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

// DeploymentDefinition represents a deployment definition
type DeploymentDefinition struct {
	Name     string   `yaml:"name"`
	Pipeline Pipeline `yaml:"pipeline"`
}

// DeploymentPlan represents a deployment plan
type DeploymentPlan struct {
	Server     string `yaml:"server"`
	Deployment string `yaml:"deployment"`
}

// Config represents the complete configuration
type Config struct {
	Servers     []Server              `yaml:"servers"`
	Deployments []DeploymentDefinition `yaml:"deployments"`
	DeployPlans []DeploymentPlan      `yaml:"deploy_plans"`
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

	// Create maps for validation
	serverNames := make(map[string]bool)
	deploymentNames := make(map[string]bool)

	// Validate servers
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

	// Validate deployments
	for _, deployment := range cfg.Deployments {
		if deployment.Name == "" {
			return fmt.Errorf("deployment name cannot be empty")
		}
		if deploymentNames[deployment.Name] {
			return fmt.Errorf("duplicate deployment name: %s", deployment.Name)
		}
		deploymentNames[deployment.Name] = true

		if len(deployment.Pipeline.Commands) == 0 {
			return fmt.Errorf("no commands defined for deployment: %s", deployment.Name)
		}
	}

	// Validate deploy plans
	for _, plan := range cfg.DeployPlans {
		if !serverNames[plan.Server] {
			return fmt.Errorf("server '%s' in deploy plan not found in server definitions", plan.Server)
		}
		if !deploymentNames[plan.Deployment] {
			return fmt.Errorf("deployment '%s' in deploy plan not found in deployment definitions", plan.Deployment)
		}
	}

	return nil
}
