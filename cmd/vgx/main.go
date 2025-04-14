package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/open-xyz/vgx/pkg/context"
	"github.com/open-xyz/vgx/pkg/scanner"
	"github.com/open-xyz/vgx/pkg/types"
	"github.com/open-xyz/vgx/pkg/vibe"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Note: No .env file found - using system environment variables")
	}

	// Parse command-line arguments
	changesOnly := flag.Bool("changes", true, "Scan only changed files")
	generateReport := flag.Bool("report", true, "Generate a report after scanning")
	updateContext := flag.Bool("update-context", true, "Update the codebase context after scanning")
	flag.Parse()

	// Initialize the context manager
	contextManager, err := context.NewContextManager()
	if err != nil {
		fmt.Printf("Error initializing context manager: %v\n", err)
		os.Exit(1)
	}

	// Clean up old reports (older than 30 days)
	if err := contextManager.CleanupOldReports(30 * 24 * time.Hour); err != nil {
		fmt.Printf("Warning: Failed to clean up old reports: %v\n", err)
	}

	// Load VibePenTester configuration
	vibeConfig := vibe.LoadConfig()
	if vibeConfig.Enabled {
		fmt.Println("🔍 VibePenTester integration enabled")
	}

	// Determine files to scan
	var filesToScan []string
	if flag.NArg() > 0 {
		// Scan specific files provided as arguments
		filesToScan = flag.Args()
	} else if *changesOnly {
		// Scan only changed files
		files, err := contextManager.GetChangedFiles()
		if err != nil {
			fmt.Printf("Error getting changed files: %v\n", err)
			os.Exit(1)
		}
		
		if len(files) == 0 {
			fmt.Println("No changed files found.")
			os.Exit(0)
		}
		
		filesToScan = files
	} else {
		// Scan all tracked files in the repository
		files, err := getAllFiles()
		if err != nil {
			fmt.Printf("Error getting all files: %v\n", err)
			os.Exit(1)
		}
		
		filesToScan = files
	}

	// Scan the files
	var allVulnerabilities []types.Vulnerability
	var scannedFiles []string

	for _, file := range filesToScan {
		// Skip directories
		info, err := os.Stat(file)
		if err != nil || info.IsDir() {
			continue
		}

		// Skip non-text files
		if !isTextFile(file) {
			continue
		}

		fmt.Printf("Scanning %s...\n", file)
		
		// Read file content
		content, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("Error: Cannot read file %s: %v\n", file, err)
			continue
		}

		// Check if file has changed since last scan
		contentStr := string(content)
		if !contextManager.HasFileChanged(file, contentStr) {
			fmt.Printf("Skipping unchanged file: %s\n", file)
			continue
		}

		scannedFiles = append(scannedFiles, file)
		
		// Get related files for context
		relatedFiles := contextManager.GetRelatedFiles(file, 5)
		var contextContent []string
		for _, relatedFile := range relatedFiles {
			if fileContext, exists := contextManager.GetFileContext(relatedFile); exists {
				contextContent = append(contextContent, fmt.Sprintf("File: %s\n%s", relatedFile, fileContext.Content))
			}
		}

		// Scan with both local and VibePenTester
		vulnerabilities := scanFilesWithAllScanners(file, contentStr, vibeConfig, contextContent)
		allVulnerabilities = append(allVulnerabilities, vulnerabilities...)
		
		// Update the context
		if *updateContext {
			if err := contextManager.UpdateFileContext(file, contentStr); err != nil {
				fmt.Printf("Warning: Failed to update context for %s: %v\n", file, err)
			}
		}
	}

	// Generate report if requested
	if *generateReport && len(scannedFiles) > 0 {
		vulnerabilityMaps := make([]map[string]interface{}, 0, len(allVulnerabilities))
		for _, vuln := range allVulnerabilities {
			vulnerabilityMaps = append(vulnerabilityMaps, map[string]interface{}{
				"file":        vuln.File,
				"line":        vuln.Line,
				"description": vuln.Description,
				"severity":    vuln.Severity,
				"source":      vuln.Source,
				"recommendation": vuln.Recommendation,
			})
		}
		
		if err := contextManager.GenerateReport(vulnerabilityMaps, scannedFiles); err != nil {
			fmt.Printf("Error generating report: %v\n", err)
		}
	}

	// Output the scan results
	if len(allVulnerabilities) > 0 {
		fmt.Println("\n🚨 VGX blocked commit due to vulnerabilities:")
		for _, vuln := range allVulnerabilities {
			fmt.Printf("  • [%s] %s (Source: %s)\n", vuln.File, vuln.Description, getSourceName(vuln.Source))
		}

		fmt.Println("\n🔧 Recommendations:")
		fmt.Println("  1. Review the flagged code")
		fmt.Println("  2. Fix the identified security issues")
		fmt.Println("  3. Commit again after resolving issues")
		
		os.Exit(1)
	} else if len(scannedFiles) > 0 {
		fmt.Println("\n✅ No security issues found in the scanned files!")
	}
}

// scanFilesWithAllScanners scans a file with all available scanners
func scanFilesWithAllScanners(filePath, content string, vibeConfig vibe.VibePenTesterConfig, contextContent []string) []types.Vulnerability {
	var allVulnerabilities []types.Vulnerability

	// Skip semgrep scanning if it's not available and avoid showing the error messages
	scanner.SetSkipSemgrepErrors(true)

	// Scan with contextual scanner (OpenAI)
	contextualVulns, err := scanner.ScanWithContext(content, filePath, contextContent)
	if err != nil {
		fmt.Printf("Warning: Contextual scan failed for %s: %v\n", filePath, err)
		// Fall back to basic scan if contextual scan fails
		localVulns, err := scanner.ScanContent(content, filePath)
		if err != nil {
			fmt.Printf("Warning: Local scan failed for %s: %v\n", filePath, err)
		} else if localVulns != nil && len(localVulns) > 0 {
			allVulnerabilities = append(allVulnerabilities, localVulns...)
		}
	} else if contextualVulns != nil && len(contextualVulns) > 0 {
		allVulnerabilities = append(allVulnerabilities, contextualVulns...)
	}

	// Scan with VibePenTester if enabled
	if vibeConfig.Enabled {
		// Add context information to the scan
		fullContent := content
		if len(contextContent) > 0 {
			fullContent = fmt.Sprintf("%s\n\nContext:\n%s", content, strings.Join(contextContent, "\n\n"))
		}
		
		vibeVulns, err := vibe.ScanContent(fullContent, filePath, vibeConfig)
		if err != nil {
			fmt.Printf("Warning: VibePenTester scan failed for %s: %v\n", filePath, err)
		} else if vibeVulns != nil && len(vibeVulns) > 0 {
			allVulnerabilities = append(allVulnerabilities, vibeVulns...)
		}
	}

	return allVulnerabilities
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

// isTextFile checks if a file is likely to be a text file
func isTextFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// List of common text file extensions to scan
	textExtensions := map[string]bool{
		".txt":  true,
		".md":   true,
		".js":   true,
		".jsx":  true,
		".ts":   true,
		".tsx":  true,
		".py":   true,
		".java": true,
		".go":   true,
		".c":    true,
		".cpp":  true,
		".h":    true,
		".hpp":  true,
		".cs":   true,
		".php":  true,
		".rb":   true,
		".html": true,
		".htm":  true,
		".css":  true,
		".scss": true,
		".json": true,
		".xml":  true,
		".yaml": true,
		".yml":  true,
		".sh":   true,
		".bash": true,
		".sql":  true,
	}
	
	return textExtensions[ext]
}

// getAllFiles gets all files in the repository
func getAllFiles() ([]string, error) {
	var files []string
	
	// Use git to list all tracked files
	cmd := exec.Command("git", "ls-files")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %v", err)
	}
	
	// Split the output into lines
	gitFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, file := range gitFiles {
		if file != "" {
			files = append(files, file)
		}
	}
	
	return files, nil
}