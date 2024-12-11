package main

import (
	"flag"
	"fmt"
	"os"
	"quick-deploy/config"
	"quick-deploy/executor"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "deploy.yaml", "Path to deployment configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create a map of servers for easy lookup
	serverMap := make(map[string]*config.Server)
	for _, server := range cfg.Servers {
		serverMap[server.Name] = &server
	}

	// Execute deployments
	for _, deployment := range cfg.Deployments.Servers {
		server, ok := serverMap[deployment.Name]
		if !ok {
			fmt.Fprintf(os.Stderr, "Server '%s' not found in configuration\n", deployment.Name)
			continue
		}

		fmt.Printf("Deploying to server: %s (%s)\n", server.Name, server.Host)
		
		// Create executor for this server
		exec, err := executor.NewExecutor(server, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating executor for server %s: %v\n", server.Name, err)
			continue
		}

		// Execute deployment pipeline
		if err := exec.ExecutePipeline(&deployment.Pipe); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing pipeline for server %s: %v\n", server.Name, err)
			continue
		}

		fmt.Printf("Deployment completed successfully for server: %s\n", server.Name)
	}
}
