package vibe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/open-xyz/vgx/pkg/scanner"
	"github.com/open-xyz/vgx/pkg/types"
)

// VibePenTesterConfig holds configuration for VibePenTester integration
type VibePenTesterConfig struct {
	Enabled    bool
	ServerURL  string
	APIKey     string
	Timeout    time.Duration
	ScanScope  string // url, domain, subdomain
	UploadLogs bool
}

// Default configuration values
const (
	DefaultServerURL = "http://localhost:5050"
	DefaultTimeout   = 60 * time.Second
	DefaultScanScope = "url"
)

// LoadConfig loads the VibePenTester configuration from environment variables
func LoadConfig() VibePenTesterConfig {
	config := VibePenTesterConfig{
		ServerURL:  getEnvOrDefault("VIBE_SERVER_URL", DefaultServerURL),
		Timeout:    time.Duration(getEnvIntOrDefault("VIBE_TIMEOUT_SECONDS", int(DefaultTimeout.Seconds()))) * time.Second,
		ScanScope:  getEnvOrDefault("VIBE_SCAN_SCOPE", DefaultScanScope),
		UploadLogs: getEnvBoolOrDefault("VIBE_UPLOAD_LOGS", false),
	}

	// Check for API key and enabled flag
	config.APIKey = os.Getenv("VIBE_API_KEY")
	config.Enabled = config.APIKey != "" && getEnvBoolOrDefault("VIBE_ENABLED", false)

	return config
}

// ScanFile sends a file to VibePenTester for additional security analysis
func ScanFile(filePath string, config VibePenTesterConfig) ([]types.Vulnerability, error) {
	if !config.Enabled {
		return nil, nil
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ScanContent(string(content), filePath, config)
}

// ScanContent sends code content to VibePenTester for security analysis
func ScanContent(content string, identifier string, config VibePenTesterConfig) ([]types.Vulnerability, error) {
	if !config.Enabled {
		return nil, nil
	}

	// First try local scanning
	vulnerabilities, err := scanner.ScanContent(content, identifier)
	if err != nil {
		return nil, err
	}

	// Then try VibePenTester remote scanning if accessible
	remoteVulns, err := sendToVibePenTester(content, identifier, config)
	if err != nil {
		fmt.Printf("Warning: VibePenTester remote scan failed: %v\n", err)
	} else if remoteVulns != nil {
		vulnerabilities = append(vulnerabilities, remoteVulns...)
	}

	return vulnerabilities, nil
}

// sendToVibePenTester sends code to the VibePenTester service for analysis
func sendToVibePenTester(content string, identifier string, config VibePenTesterConfig) ([]types.Vulnerability, error) {
	client := &http.Client{
		Timeout: config.Timeout,
	}

	// Create request body
	type requestBody struct {
		Code       string `json:"code"`
		Identifier string `json:"identifier"`
		ScanScope  string `json:"scan_scope"`
	}

	body, err := json.Marshal(requestBody{
		Code:       content,
		Identifier: identifier,
		ScanScope:  config.ScanScope,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/scan", config.ServerURL), strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if config.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	type responseBody struct {
		Success       bool `json:"success"`
		Vulnerabilities []struct {
			Type        string `json:"type"`
			Description string `json:"description"`
			Severity    string `json:"severity"`
			Line        int    `json:"line"`
		} `json:"vulnerabilities"`
	}

	var response responseBody
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	// Convert response to Vulnerability structs
	var vulnerabilities []types.Vulnerability
	for _, v := range response.Vulnerabilities {
		vulnerabilities = append(vulnerabilities, types.Vulnerability{
			File:        identifier,
			Description: v.Description,
			Severity:    v.Severity,
			Line:        v.Line,
			Source:      "vibepentester",
		})
	}

	return vulnerabilities, nil
}

// Helper function to get environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Helper function to get int environment variable or default value
func getEnvIntOrDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err != nil {
		return defaultValue
	}
	return result
}

// Helper function to get bool environment variable or default value
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return strings.ToLower(value) == "true" || value == "1"
} 