package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/open-xyz/vgx/pkg/git"
	"github.com/open-xyz/vgx/pkg/scanner"
	"github.com/open-xyz/vgx/pkg/types"
	"github.com/open-xyz/vgx/pkg/vibe"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Note: No .env file found - using system environment variables")
	}

	// Get staged files
	files, err := git.GetStagedFiles()
	if err != nil {
		fmt.Printf("Error getting staged files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("✅ No files to scan - commit allowed")
		os.Exit(0)
	}

	// Load VibePenTester configuration
	vibeConfig := vibe.LoadConfig()
	if vibeConfig.Enabled {
		fmt.Println("🔍 VibePenTester integration enabled")
	}

	// Scan files
	vulnerabilities, err := scanFilesWithAllScanners(files, vibeConfig)
	if err != nil {
		fmt.Printf("Scan failed: %v\n", err)
		os.Exit(1)
	}

	// Block commit if vulnerabilities found
	if len(vulnerabilities) > 0 {
		fmt.Println("🚨 VGX blocked commit due to vulnerabilities:")
		for _, v := range vulnerabilities {
			fmt.Printf("  • [%s] %s (Source: %s)\n", v.File, v.Description, getSourceName(v.Source))
		}
		fmt.Println("\n🔧 Recommendations:")
		fmt.Println("  1. Review the flagged code")
		fmt.Println("  2. Fix the identified security issues")
		fmt.Println("  3. Commit again after resolving issues")
		os.Exit(1)
	}

	fmt.Println("✅ VGX: No vulnerabilities found - commit allowed!")
}

// scanFilesWithAllScanners scans files using all available scanners
func scanFilesWithAllScanners(files []string, vibeConfig vibe.VibePenTesterConfig) ([]types.Vulnerability, error) {
	// First run the standard scanning
	vulnerabilities, err := scanner.ScanFiles(files)
	if err != nil {
		return nil, err
	}

	// If VibePenTester integration is enabled, also scan files with that
	if vibeConfig.Enabled {
		for _, file := range files {
			vibeVulns, err := vibe.ScanFile(file, vibeConfig)
			if err != nil {
				fmt.Printf("Warning: VibePenTester scan failed for %s: %v\n", file, err)
			} else if vibeVulns != nil && len(vibeVulns) > 0 {
				vulnerabilities = append(vulnerabilities, vibeVulns...)
			}
		}
	}

	return vulnerabilities, nil
}

// getSourceName returns a user-friendly name for the scan source
func getSourceName(source string) string {
	switch source {
	case "semgrep":
		return "Semgrep"
	case "openai":
		return "OpenAI"
	case "vibepentester":
		return "VibePenTester"
	default:
		return source
	}
}