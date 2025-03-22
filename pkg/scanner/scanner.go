package scanner

import (
	"fmt"
)

// ScanFiles scans multiple files for security vulnerabilities
func ScanFiles(files []string) ([]Vulnerability, error) {
	var allVulnerabilities []Vulnerability

	for _, file := range files {
		// Run Semgrep on each file
		vulns, err := RunSemgrep(file)
		if err != nil {
			return nil, fmt.Errorf("semgrep scan failed on %s: %w", file, err)
		}
		
		allVulnerabilities = append(allVulnerabilities, vulns...)

		// Optional: Also analyze with OpenAI if configured
		if len(vulns) == 0 {
			// Only if Semgrep didn't find anything, use OpenAI as second opinion
			// Commented out additional OpenAI analysis code
		}
	}

	return allVulnerabilities, nil
} 