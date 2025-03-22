package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/open-xyz/vibe-guard/pkg/git"
	"github.com/open-xyz/vibe-guard/pkg/scanner"
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

	// Scan files
	vulnerabilities, err := scanner.ScanFiles(files)
	if err != nil {
		fmt.Printf("Scan failed: %v\n", err)
		os.Exit(1)
	}

	// Block commit if vulnerabilities found
	if len(vulnerabilities) > 0 {
		fmt.Println("🚨 VibeGuard blocked commit due to vulnerabilities:")
		for _, v := range vulnerabilities {
			fmt.Printf("  • [%s] %s\n", v.File, v.Description)
		}
		fmt.Println("\n🔧 Recommendations:")
		fmt.Println("  1. Review the flagged code")
		fmt.Println("  2. Use 'vibeguard fix <file>' for auto-fixes")
		fmt.Println("  3. Commit again after resolving issues")
		os.Exit(1)
	}

	fmt.Println("✅ VibeGuard: No vulnerabilities found - commit allowed!")
}