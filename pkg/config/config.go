package config

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

type Resource struct {
	DomainName string `yaml:"DomainName" validate:"required"`
	Port       int    `yaml:"Port" validate:"required,min=1,max=65535"`
	Endpoint   string `yaml:"Endpoint" validate:"required"`
}

type Proxy struct {
	DomainName string     `yaml:"DomainName" validate:"required"`
	Port       int        `yaml:"Port" validate:"required,min=1,max=65535"`
	Resources  []Resource `yaml:"Resources" validate:"required,min=1"`
}

// ConfigError represents configuration-related errors
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config validation error in field '%s': %s", e.Field, e.Message)
}

// LoadConfig loads configuration from a YAML file with validation
func LoadConfig(filename string) (*Proxy, error) {
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file '%s' does not exist", filename)
	}

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file '%s': %w", filename, err)
	}

	// Parse YAML
	var config Proxy
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfigFromBytes loads configuration from byte slice with validation
func LoadConfigFromBytes(data []byte) (*Proxy, error) {
	var config Proxy
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateConfig performs comprehensive validation on the configuration
func validateConfig(config *Proxy) error {
	// Validate main proxy configuration
	if config.DomainName == "" {
		return &ConfigError{Field: "DomainName", Message: "domain name is required"}
	}

	if config.Port < 1 || config.Port > 65535 {
		return &ConfigError{Field: "Port", Message: "port must be between 1 and 65535"}
	}

	// Validate resources
	if len(config.Resources) == 0 {
		return &ConfigError{Field: "Resources", Message: "at least one resource is required"}
	}

	for i, resource := range config.Resources {
		if err := validateResource(&resource, i); err != nil {
			return err
		}
	}

	return nil
}

// validateResource validates individual resource configuration
func validateResource(resource *Resource, index int) error {
	fieldPrefix := fmt.Sprintf("Resources[%d]", index)

	if resource.DomainName == "" {
		return &ConfigError{
			Field:   fmt.Sprintf("%s.DomainName", fieldPrefix),
			Message: "domain name is required",
		}
	}

	if resource.Endpoint == "" {
		return &ConfigError{
			Field:   fmt.Sprintf("%s.Endpoint", fieldPrefix),
			Message: "Endpoint is required",
		}
	}

	if resource.Port < 1 || resource.Port > 65535 {
		return &ConfigError{
			Field:   fmt.Sprintf("%s.Port", fieldPrefix),
			Message: "port must be between 1 and 65535",
		}
	}

	return nil
}

// SaveConfig saves configuration to a YAML file
func SaveConfig(config *Proxy, filename string) error {
	// Validate before saving
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("cannot save invalid config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dir, err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file '%s': %w", filename, err)
	}

	return nil
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Proxy {
	return &Proxy{
		DomainName: "localhost",
		Port:       8080,
		Resources: []Resource{
			{
				DomainName: "localhost",
				Port:       8080,
			},
		},
	}
}
