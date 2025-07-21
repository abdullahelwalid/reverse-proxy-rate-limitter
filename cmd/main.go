package main

import (
	"fmt"
	"log"
	"os"

	"github.com/abdullahelwalid/go-rate-limiter/pkg/config"
	"github.com/abdullahelwalid/go-rate-limiter/pkg/limitter"
	"github.com/abdullahelwalid/go-rate-limiter/pkg/server"
)

func main() {
	// Load configuration from file
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Load Redis Client Connection
	limitter.GetClient()

	fmt.Printf("Loaded config:\n")
	fmt.Printf("  Domain: %s\n", cfg.DomainName)
	fmt.Printf("  Port: %d\n", cfg.Port)
	fmt.Printf("  Resources: %d\n", len(cfg.Resources))
	for i, resource := range cfg.Resources {
		fmt.Printf("    Resource %d: %s:%d\n", i+1, resource.DomainName, resource.Port)
	}
	server.RunServer(*cfg)
}

func createTestFiles() {
	// Create an invalid YAML file for testing
	invalidYAML := `
DomainName: localhost
Port: 99999  # Invalid port number
Resources:
  - DomainName: ""  # Empty domain name
    Port: 8080
`

	if err := os.WriteFile("invalid.yaml", []byte(invalidYAML), 0644); err != nil {
		log.Printf("Failed to create test file: %v", err)
	}

	// Create a valid config file
	validYAML := `
DomainName: localhost
Port: 8080
Resources:
  - DomainName: localhost
    Port: 8080
  - DomainName: api.example.com
    Port: 9000
`

	if err := os.WriteFile("config.yaml", []byte(validYAML), 0644); err != nil {
		log.Printf("Failed to create config file: %v", err)
	}
}
