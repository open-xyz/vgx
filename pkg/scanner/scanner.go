package scanner

import (
	"fmt"

	"github.com/open-xyz/vgx/pkg/cache"
	"github.com/open-xyz/vgx/pkg/types"
)

// ScanFiles scans multiple files for security vulnerabilities
func ScanFiles(files []string) ([]types.Vulnerability, error) {
	var allVulnerabilities []types.Vulnerability

	for _, file := range files {
		// Try to get results from cache first
		if cachedVulns, found, err := cache.Get(file); err == nil && found {
			// Cache hit - use cached results
			fmt.Printf("Using cached results for %s\n", file)
			allVulnerabilities = append(allVulnerabilities, cachedVulns...)
			continue
		}
		
		// Cache miss - run actual scan
		vulns, err := RunSemgrep(file)
		if err != nil {
			return nil, fmt.Errorf("semgrep scan failed on %s: %w", file, err)
		}
		
		// Store results in cache for future use
		if err := cache.Store(file, vulns); err != nil {
			// Non-fatal error, just log it
			fmt.Printf("Warning: Failed to cache results for %s: %v\n", file, err)
		}
		
		allVulnerabilities = append(allVulnerabilities, vulns...)

		// Optional: Also analyze with OpenAI if configured
		if len(vulns) == 0 {
			// Only if Semgrep didn't find anything, use OpenAI as second opinion
			// Read file content
			// content, err := os.ReadFile(file)
			// if err == nil {
			//     result, err := AnalyzeWithOpenAI(string(content))
			//     if err == nil && strings.HasPrefix(result, "UNSAFE:") {
			//         vulnResult := Vulnerability{
			//             File:        file,
			//             Description: strings.TrimPrefix(result, "UNSAFE: "),
			//             Severity:    "medium",
			//             Source:      "openai",
			//         }
			//         allVulnerabilities = append(allVulnerabilities, vulnResult)
			//         
			//         // Also cache the OpenAI result
			//         if err := cache.Store(file, []Vulnerability{vulnResult}); err != nil {
			//             fmt.Printf("Warning: Failed to cache OpenAI results for %s: %v\n", file, err)
			//         }
			//     }
			// }
		}
	}

	return allVulnerabilities, nil
}