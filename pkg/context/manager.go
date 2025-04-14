package context

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// CodebaseContext represents the stored context of the codebase
type CodebaseContext struct {
	LastUpdated time.Time          `json:"last_updated"`
	Files       map[string]FileContext `json:"files"`
	Version     string             `json:"version"`
}

// FileContext stores information about a file
type FileContext struct {
	Path      string    `json:"path"`
	Hash      string    `json:"hash"`
	LastScan  time.Time `json:"last_scan"`
	Content   string    `json:"content,omitempty"`
	Embedding []float64 `json:"embedding,omitempty"`
}

// ContextManager handles the codebase context operations
type ContextManager struct {
	contextDir    string
	contextFile   string
	codebase      CodebaseContext
	isInitialized bool
}

// NewContextManager creates a new context manager
func NewContextManager() (*ContextManager, error) {
	// Load env variables
	godotenv.Load()

	// Get home directory for storing context
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	// Create context directory
	contextDir := filepath.Join(homeDir, ".vgx")
	if err := os.MkdirAll(contextDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create context directory: %v", err)
	}

	contextFile := filepath.Join(contextDir, "context.json")
	
	manager := &ContextManager{
		contextDir:  contextDir,
		contextFile: contextFile,
		codebase: CodebaseContext{
			Files:   make(map[string]FileContext),
			Version: "1.0",
		},
		isInitialized: false,
	}

	// Load existing context if available
	if _, err := os.Stat(contextFile); err == nil {
		if err := manager.loadContext(); err != nil {
			fmt.Printf("Warning: Failed to load existing context: %v\n", err)
		} else {
			manager.isInitialized = true
		}
	}

	return manager, nil
}

// loadContext loads the context from disk
func (m *ContextManager) loadContext() error {
	data, err := ioutil.ReadFile(m.contextFile)
	if err != nil {
		return fmt.Errorf("failed to read context file: %v", err)
	}

	if err := json.Unmarshal(data, &m.codebase); err != nil {
		return fmt.Errorf("failed to parse context file: %v", err)
	}

	return nil
}

// saveContext saves the context to disk
func (m *ContextManager) saveContext() error {
	data, err := json.MarshalIndent(m.codebase, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal context: %v", err)
	}

	if err := ioutil.WriteFile(m.contextFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write context file: %v", err)
	}

	return nil
}

// GetChangedFiles returns a list of files that have changed since the last scan
func (m *ContextManager) GetChangedFiles() ([]string, error) {
	// If not in a git repository, return an empty list
	if !isGitRepository() {
		return []string{}, nil
	}

	// Get list of changed files from git
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %v", err)
	}

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	
	// Filter out empty entries
	var filtered []string
	for _, file := range changedFiles {
		if file != "" {
			filtered = append(filtered, file)
		}
	}

	// If no changes in the index, check for unstaged changes
	if len(filtered) == 0 {
		cmd = exec.Command("git", "diff", "--name-only")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get unstaged changes: %v", err)
		}

		changedFiles = strings.Split(strings.TrimSpace(string(output)), "\n")
		
		for _, file := range changedFiles {
			if file != "" {
				filtered = append(filtered, file)
			}
		}
	}

	// Also include untracked files
	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get untracked files: %v", err)
	}

	untrackedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, file := range untrackedFiles {
		if file != "" {
			filtered = append(filtered, file)
		}
	}

	return filtered, nil
}

// UpdateFileContext updates the context for a specific file
func (m *ContextManager) UpdateFileContext(filePath string, content string) error {
	if !m.isInitialized {
		fmt.Println("Initializing codebase context...")
	}

	hash := getFileHash(content)
	
	m.codebase.Files[filePath] = FileContext{
		Path:     filePath,
		Hash:     hash,
		LastScan: time.Now(),
		Content:  content,
	}
	
	m.codebase.LastUpdated = time.Now()
	m.isInitialized = true
	
	return m.saveContext()
}

// HasFileChanged checks if a file has changed since the last scan
func (m *ContextManager) HasFileChanged(filePath string, content string) bool {
	if !m.isInitialized {
		return true
	}
	
	if fileContext, exists := m.codebase.Files[filePath]; exists {
		currentHash := getFileHash(content)
		return currentHash != fileContext.Hash
	}
	
	return true
}

// GetFileContext returns the context for a specific file
func (m *ContextManager) GetFileContext(filePath string) (FileContext, bool) {
	context, exists := m.codebase.Files[filePath]
	return context, exists
}

// GetRelatedFiles returns files that might be related to the given file
// This is a placeholder for future implementation with embeddings
func (m *ContextManager) GetRelatedFiles(filePath string, limit int) []string {
	// This would use embeddings to find related files
	// For now, just return all files in the same directory
	
	dir := filepath.Dir(filePath)
	var related []string
	
	for path := range m.codebase.Files {
		if filepath.Dir(path) == dir && path != filePath {
			related = append(related, path)
			if len(related) >= limit {
				break
			}
		}
	}
	
	return related
}

// isGitRepository checks if the current directory is a git repository
func isGitRepository() bool {
	_, err := os.Stat(".git")
	return err == nil
}

// getFileHash generates a simple hash for file content
func getFileHash(content string) string {
	// This is a simplified hash, in a real implementation
	// you might want to use a more robust hashing algorithm
	return fmt.Sprintf("%d", len(content))
}

// GenerateReport creates a report of the scan results
func (m *ContextManager) GenerateReport(vulnerabilities []map[string]interface{}, scannedFiles []string) error {
	if len(vulnerabilities) == 0 && len(scannedFiles) == 0 {
		return nil
	}

	reportDir := "reports"
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %v", err)
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	reportFile := filepath.Join(reportDir, fmt.Sprintf("security-report-%s.md", timestamp))

	var report strings.Builder
	report.WriteString("# VGX Security Scan Report\n\n")
	report.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC1123)))

	report.WriteString("## Files Scanned\n\n")
	for _, file := range scannedFiles {
		report.WriteString(fmt.Sprintf("- %s\n", file))
	}
	report.WriteString("\n")

	if len(vulnerabilities) > 0 {
		report.WriteString("## Vulnerabilities Found\n\n")
		
		for i, vuln := range vulnerabilities {
			report.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, vuln["description"]))
			report.WriteString(fmt.Sprintf("- **File**: %s\n", vuln["file"]))
			report.WriteString(fmt.Sprintf("- **Severity**: %s\n", vuln["severity"]))
			
			if line, ok := vuln["line"].(int); ok && line > 0 {
				report.WriteString(fmt.Sprintf("- **Line**: %d\n", line))
			}
			
			if recommendation, ok := vuln["recommendation"].(string); ok && recommendation != "" {
				report.WriteString(fmt.Sprintf("\n**Recommendation**: %s\n", recommendation))
			}
			
			report.WriteString("\n")
		}
	} else {
		report.WriteString("## No Vulnerabilities Found\n\n")
		report.WriteString("✅ Great job! No security issues were detected in the scanned files.\n\n")
	}

	if err := ioutil.WriteFile(reportFile, []byte(report.String()), 0644); err != nil {
		return fmt.Errorf("failed to write report file: %v", err)
	}

	fmt.Printf("Report generated: %s\n", reportFile)
	return nil
}

// CleanupOldReports removes reports older than a certain age
func (m *ContextManager) CleanupOldReports(maxAge time.Duration) error {
	reportDir := "reports"
	if _, err := os.Stat(reportDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := ioutil.ReadDir(reportDir)
	if err != nil {
		return fmt.Errorf("failed to read reports directory: %v", err)
	}

	now := time.Now()
	for _, entry := range entries {
		if now.Sub(entry.ModTime()) > maxAge {
			if err := os.Remove(filepath.Join(reportDir, entry.Name())); err != nil {
				fmt.Printf("Warning: Failed to remove old report %s: %v\n", entry.Name(), err)
			}
		}
	}

	return nil
} 