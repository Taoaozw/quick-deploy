package main

import (
	"flag"
	"fmt"
	"os"
	"quick-deploy/config"
	"quick-deploy/executor"
)

func executeDeploymentPlan(plan *config.DeploymentPlan, 
	server *config.Server, 
	deployment *config.DeploymentDefinition) error {
	
	fmt.Printf("\n==> Executing deployment '%s' on server: %s (%s)\n", 
		deployment.Name, server.Name, server.Host)
	
	// Create executor for this server
	exec, err := executor.NewExecutor(server, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to create executor: %v", err)
	}
	defer exec.Close()

	// Execute deployment pipeline
	if err := exec.ExecutePipeline(&deployment.Pipeline); err != nil {
		return fmt.Errorf("pipeline execution failed: %v", err)
	}

	return nil
}

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

	// Create maps for quick lookup
	serverMap := make(map[string]*config.Server)
	for _, server := range cfg.Servers {
		serverMap[server.Name] = &server
	}

	deploymentMap := make(map[string]*config.DeploymentDefinition)
	for _, deployment := range cfg.Deployments {
		deploymentMap[deployment.Name] = &deployment
	}

	// Track overall success
	success := true
	totalPlans := len(cfg.DeployPlans)
	completedPlans := 0

	fmt.Printf("Starting deployment with %d plans\n", totalPlans)

	// Execute deployment plans
	for _, plan := range cfg.DeployPlans {
		server := serverMap[plan.Server]
		deployment := deploymentMap[plan.Deployment]

		err := executeDeploymentPlan(&plan, server, deployment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error in deployment plan (server: %s, deployment: %s): %v\n",
				plan.Server, plan.Deployment, err)
			success = false
		} else {
			completedPlans++
		}
	}

	// Print final status
	fmt.Printf("\n==> Deployment Summary:\n")
	fmt.Printf("Total plans: %d\n", totalPlans)
	fmt.Printf("Completed successfully: %d\n", completedPlans)
	fmt.Printf("Failed: %d\n", totalPlans-completedPlans)

	if !success {
		os.Exit(1)
	}
}
