package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/open-xyz/vgx/pkg/scanner"
	"github.com/open-xyz/vgx/pkg/types"
	"github.com/open-xyz/vgx/pkg/vibe"
)

// Define a reusable struct for scan results
type ScanResult struct {
	Timestamp      string                `json:"timestamp"`
	FilesScanned   []string              `json:"files_scanned"`
	Vulnerabilities []types.Vulnerability `json:"vulnerabilities"`
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Note: No .env file found - using system environment variables")
	}

	// Define command flags
	scanCmd := flag.NewFlagSet("scan", flag.ExitOnError)
	scanOutput := scanCmd.String("output", "", "Output file for scan results (JSON)")
	scanFormat := scanCmd.String("format", "json", "Output format (json, text)")
	scanRecursive := scanCmd.Bool("recursive", false, "Scan directories recursively")
	
	reportCmd := flag.NewFlagSet("report", flag.ExitOnError)
	reportInput := reportCmd.String("input", "", "Input JSON file with scan results")
	reportOutput := reportCmd.String("output", "", "Output file for report (HTML, PDF)")
	reportFormat := reportCmd.String("format", "html", "Output format (html, pdf, markdown)")

	// Check if we have enough arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse the appropriate command
	switch os.Args[1] {
	case "scan":
		scanCmd.Parse(os.Args[2:])
		runScan(scanCmd.Args(), *scanOutput, *scanFormat, *scanRecursive)
	case "report":
		reportCmd.Parse(os.Args[2:])
		generateReport(*reportInput, *reportOutput, *reportFormat)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("VGX VibePenTester Integration")
	fmt.Println("\nUsage:")
	fmt.Println("  vibe scan [flags] [files/directories...]   - Scan files for vulnerabilities")
	fmt.Println("  vibe report [flags]                        - Generate report from scan results")
	fmt.Println("\nScan flags:")
	fmt.Println("  -output string    Output file for scan results (JSON)")
	fmt.Println("  -format string    Output format (json, text) (default \"json\")")
	fmt.Println("  -recursive        Scan directories recursively")
	fmt.Println("\nReport flags:")
	fmt.Println("  -input string     Input JSON file with scan results")
	fmt.Println("  -output string    Output file for report (HTML, PDF)")
	fmt.Println("  -format string    Output format (html, pdf, markdown) (default \"html\")")
}

func runScan(args []string, outputFile, format string, recursive bool) {
	// Check if we have files to scan
	if len(args) == 0 {
		fmt.Println("Error: No files or directories specified")
		os.Exit(1)
	}

	// Load VibePenTester configuration
	vibeConfig := vibe.LoadConfig()
	if vibeConfig.Enabled {
		fmt.Println("🔍 VibePenTester integration enabled")
	}

	// Collect all files to scan
	var filesToScan []string
	for _, arg := range args {
		fileInfo, err := os.Stat(arg)
		if err != nil {
			fmt.Printf("Error: Cannot access %s: %v\n", arg, err)
			continue
		}

		if fileInfo.IsDir() {
			if recursive {
				// Recursively collect files from directory
				filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						filesToScan = append(filesToScan, path)
					}
					return nil
				})
			} else {
				// Read only files in the top directory
				files, err := ioutil.ReadDir(arg)
				if err != nil {
					fmt.Printf("Error: Cannot read directory %s: %v\n", arg, err)
					continue
				}

				for _, f := range files {
					if !f.IsDir() {
						filesToScan = append(filesToScan, filepath.Join(arg, f.Name()))
					}
				}
			}
		} else {
			// Add individual file
			filesToScan = append(filesToScan, arg)
		}
	}

	// Scan files
	var allVulnerabilities []types.Vulnerability
	for _, file := range filesToScan {
		fmt.Printf("Scanning %s...\n", file)
		
		// Read file content
		content, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Printf("Error: Cannot read file %s: %v\n", file, err)
			continue
		}

		// Scan with OpenAI
		openaiVulns, err := scanner.ScanContent(string(content), file)
		if err != nil {
			fmt.Printf("Warning: Local scan failed for %s: %v\n", file, err)
		} else if openaiVulns != nil && len(openaiVulns) > 0 {
			allVulnerabilities = append(allVulnerabilities, openaiVulns...)
		}

		// Scan with VibePenTester if enabled
		if vibeConfig.Enabled {
			vibeVulns, err := vibe.ScanContent(string(content), file, vibeConfig)
			if err != nil {
				fmt.Printf("Warning: VibePenTester scan failed for %s: %v\n", file, err)
			} else if vibeVulns != nil && len(vibeVulns) > 0 {
				allVulnerabilities = append(allVulnerabilities, vibeVulns...)
			}
		}
	}

	// Output results
	if format == "json" || outputFile != "" {
		// Create result structure
		result := ScanResult{
			Timestamp:      time.Now().Format(time.RFC3339),
			FilesScanned:   filesToScan,
			Vulnerabilities: allVulnerabilities,
		}

		// Marshal to JSON
		jsonOutput, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("Error: Failed to marshal JSON: %v\n", err)
			os.Exit(1)
		}

		// Write to file if specified
		if outputFile != "" {
			err := ioutil.WriteFile(outputFile, jsonOutput, 0644)
			if err != nil {
				fmt.Printf("Error: Failed to write output file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Results written to %s\n", outputFile)
		}

		// Print to stdout if format is json
		if format == "json" && outputFile == "" {
			fmt.Println(string(jsonOutput))
		}
	}

	// Print text output
	if format == "text" || outputFile != "" {
		fmt.Printf("\n=== VGX Scan Results ===\n")
		fmt.Printf("Scanned %d files\n", len(filesToScan))
		fmt.Printf("Found %d vulnerabilities\n\n", len(allVulnerabilities))

		if len(allVulnerabilities) > 0 {
			for i, v := range allVulnerabilities {
				fmt.Printf("Vulnerability #%d:\n", i+1)
				fmt.Printf("  File: %s\n", v.File)
				fmt.Printf("  Description: %s\n", v.Description)
				fmt.Printf("  Severity: %s\n", v.Severity)
				if v.Line > 0 {
					fmt.Printf("  Line: %d\n", v.Line)
				}
				fmt.Printf("  Source: %s\n\n", getSourceName(v.Source))
			}
		} else {
			fmt.Println("No vulnerabilities found!")
		}
	}

	// Exit with code 1 if vulnerabilities found
	if len(allVulnerabilities) > 0 {
		os.Exit(1)
	}
}

func generateReport(inputFile, outputFile, format string) {
	// Check if input file is specified
	if inputFile == "" {
		fmt.Println("Error: Input file must be specified")
		os.Exit(1)
	}

	// Read input file
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		fmt.Printf("Error: Cannot read input file: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON
	var scanResult ScanResult
	if err := json.Unmarshal(data, &scanResult); err != nil {
		fmt.Printf("Error: Cannot parse JSON: %v\n", err)
		os.Exit(1)
	}

	// Generate report
	var reportContent string
	switch format {
	case "html":
		reportContent = generateHtmlReport(scanResult)
	case "markdown":
		reportContent = generateMarkdownReport(scanResult)
	case "pdf":
		fmt.Println("Error: PDF output not yet implemented")
		os.Exit(1)
	default:
		fmt.Printf("Error: Unknown format: %s\n", format)
		os.Exit(1)
	}

	// Write to file or stdout
	if outputFile != "" {
		if err := ioutil.WriteFile(outputFile, []byte(reportContent), 0644); err != nil {
			fmt.Printf("Error: Cannot write output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Report written to %s\n", outputFile)
	} else {
		fmt.Println(reportContent)
	}
}

func generateHtmlReport(result ScanResult) string {
	// Simple HTML report template
	template := `<!DOCTYPE html>
<html>
<head>
    <title>VGX Security Report</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; margin: 0 auto; max-width: 1000px; padding: 20px; }
        h1, h2, h3 { color: #333; }
        .summary { background-color: #f5f5f5; border-radius: 5px; padding: 15px; margin-bottom: 20px; }
        .vulnerability { background-color: #fff8f8; border-left: 4px solid #e74c3c; padding: 15px; margin-bottom: 15px; }
        .high { border-color: #e74c3c; }
        .medium { border-color: #f39c12; }
        .low { border-color: #3498db; }
        .file-list { color: #666; font-size: 0.9em; max-height: 200px; overflow-y: auto; }
        .timestamp { color: #888; font-size: 0.9em; }
        table { border-collapse: collapse; width: 100%; }
        th, td { text-align: left; padding: 8px; border-bottom: 1px solid #ddd; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <h1>VGX Security Report</h1>
    <div class="timestamp">Generated: %s</div>
    
    <div class="summary">
        <h2>Summary</h2>
        <p>Scanned %d files and found %d vulnerabilities</p>
    </div>
    
    <h2>Vulnerabilities</h2>
    %s
    
    <h2>Scanned Files</h2>
    <div class="file-list">
        <ul>
            %s
        </ul>
    </div>
</body>
</html>`

	// Generate vulnerabilities HTML
	var vulnsHtml strings.Builder
	for _, v := range result.Vulnerabilities {
		severity := strings.ToLower(v.Severity)
		vulnsHtml.WriteString(fmt.Sprintf(`
        <div class="vulnerability %s">
            <h3>%s</h3>
            <table>
                <tr>
                    <th>File</th>
                    <td>%s</td>
                </tr>
                <tr>
                    <th>Severity</th>
                    <td>%s</td>
                </tr>`,
			severity, v.Description, v.File, v.Severity))

		if v.Line > 0 {
			vulnsHtml.WriteString(fmt.Sprintf(`
                <tr>
                    <th>Line</th>
                    <td>%d</td>
                </tr>`, v.Line))
		}

		vulnsHtml.WriteString(fmt.Sprintf(`
                <tr>
                    <th>Source</th>
                    <td>%s</td>
                </tr>
            </table>
        </div>`, getSourceName(v.Source)))
	}

	// Generate file list HTML
	var filesHtml strings.Builder
	for _, file := range result.FilesScanned {
		filesHtml.WriteString(fmt.Sprintf("<li>%s</li>\n", file))
	}

	// Format the report
	return fmt.Sprintf(template,
		result.Timestamp,
		len(result.FilesScanned),
		len(result.Vulnerabilities),
		vulnsHtml.String(),
		filesHtml.String())
}

func generateMarkdownReport(result ScanResult) string {
	// Simple Markdown report template
	var report strings.Builder

	report.WriteString("# VGX Security Report\n\n")
	report.WriteString(fmt.Sprintf("Generated: %s\n\n", result.Timestamp))

	report.WriteString("## Summary\n\n")
	report.WriteString(fmt.Sprintf("Scanned %d files and found %d vulnerabilities\n\n", 
		len(result.FilesScanned), len(result.Vulnerabilities)))

	if len(result.Vulnerabilities) > 0 {
		report.WriteString("## Vulnerabilities\n\n")

		for i, v := range result.Vulnerabilities {
			report.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, v.Description))
			report.WriteString(fmt.Sprintf("- **File**: %s\n", v.File))
			report.WriteString(fmt.Sprintf("- **Severity**: %s\n", v.Severity))
			
			if v.Line > 0 {
				report.WriteString(fmt.Sprintf("- **Line**: %d\n", v.Line))
			}
			
			report.WriteString(fmt.Sprintf("- **Source**: %s\n\n", getSourceName(v.Source)))
		}
	}

	report.WriteString("## Scanned Files\n\n")
	for _, file := range result.FilesScanned {
		report.WriteString(fmt.Sprintf("- %s\n", file))
	}

	return report.String()
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