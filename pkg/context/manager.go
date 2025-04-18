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
	report.WriteString("# 🛡️ VGX Security Scan Report\n\n")
	report.WriteString(fmt.Sprintf("📅 Generated: %s\n\n", time.Now().Format(time.RFC1123)))

	report.WriteString("## 📁 Files Scanned\n\n")
	for _, file := range scannedFiles {
		report.WriteString(fmt.Sprintf("- %s\n", file))
	}
	report.WriteString("\n")

	if len(vulnerabilities) > 0 {
		report.WriteString("## 🚨 Vulnerabilities Found\n\n")
		
		for i, vuln := range vulnerabilities {
			// Add emoji based on severity
			severityEmoji := "⚠️" // Default (medium)
			if sev, ok := vuln["severity"].(string); ok {
				switch strings.ToLower(sev) {
				case "critical":
					severityEmoji = "💥"
				case "high":
					severityEmoji = "🔴"
				case "medium":
					severityEmoji = "🟠"
				case "low":
					severityEmoji = "🟡"
				case "info":
					severityEmoji = "🔵"
				}
			}
			
			description := ""
			if desc, ok := vuln["description"].(string); ok {
				description = desc
			}
			
			report.WriteString(fmt.Sprintf("### %d. %s %s\n\n", i+1, severityEmoji, description))
			report.WriteString(fmt.Sprintf("- **File**: %s\n", vuln["file"]))
			report.WriteString(fmt.Sprintf("- **Severity**: %s\n", vuln["severity"]))
			
			line := 0
			if lineVal, ok := vuln["line"].(int); ok && lineVal > 0 {
				line = lineVal
				report.WriteString(fmt.Sprintf("- **Line**: %d\n", line))
			}
			
			// Try to get the code snippet if the file exists
			if filePath, ok := vuln["file"].(string); ok && line > 0 {
				content, err := m.getCodeSnippet(filePath, line, 5)
				if err == nil {
					report.WriteString("\n**Vulnerable Code**:\n")
					report.WriteString("```javascript\n")
					report.WriteString(content)
					report.WriteString("\n```\n")
				}
			}
			
			if recommendation, ok := vuln["recommendation"].(string); ok && recommendation != "" {
				report.WriteString(fmt.Sprintf("\n**💡 Recommendation**: %s\n", recommendation))
			}
			
			// Add corrected code example based on vulnerability type
			if desc, ok := vuln["description"].(string); ok {
				correctedCode := m.getFixedCodeExample(desc, vuln)
				if correctedCode != "" {
					report.WriteString("\n**✅ Fixed Code Example**:\n")
					report.WriteString("```javascript\n")
					report.WriteString(correctedCode)
					report.WriteString("\n```\n")
				}
			}
			
			report.WriteString("\n")
		}
	} else {
		report.WriteString("## ✅ No Vulnerabilities Found\n\n")
		report.WriteString("🎉 Great job! No security issues were detected in the scanned files.\n\n")
	}

	if err := ioutil.WriteFile(reportFile, []byte(report.String()), 0644); err != nil {
		return fmt.Errorf("failed to write report file: %v", err)
	}

	fmt.Printf("Report generated: %s\n", reportFile)
	return nil
}

// getCodeSnippet retrieves a code snippet from a file centered around the specified line
func (m *ContextManager) getCodeSnippet(filePath string, line, context int) (string, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	
	content := string(bytes)
	lines := strings.Split(content, "\n")
	
	if line > len(lines) {
		return "", fmt.Errorf("line number out of range")
	}
	
	start := line - context
	if start < 0 {
		start = 0
	}
	
	end := line + context
	if end > len(lines) {
		end = len(lines)
	}
	
	var snippet strings.Builder
	for i := start; i < end; i++ {
		if i+1 == line {
			snippet.WriteString(fmt.Sprintf("➡️ %d: %s\n", i+1, lines[i]))
		} else {
			snippet.WriteString(fmt.Sprintf("   %d: %s\n", i+1, lines[i]))
		}
	}
	
	return snippet.String(), nil
}

// getFixedCodeExample provides a corrected code example for a specific vulnerability type
func (m *ContextManager) getFixedCodeExample(description string, vuln map[string]interface{}) string {
	// Parse the description to determine the vulnerability type
	descLower := strings.ToLower(description)
	
	if strings.Contains(descLower, "information disclosure") && strings.Contains(descLower, "error logging") {
		return `// Create a safe error logging function
const logger = {
  error: (message, error) => {
    // Only log non-sensitive information
    console.error(`+"`"+`[ERROR] ${message}: ${error.message || 'Unknown error'}`+"`"+`);
    
    // For debugging in development only
    if (process.env.NODE_ENV === 'development') {
      console.debug('Error details:', error);
    }
  }
};`
	} else if strings.Contains(descLower, "jwt") && strings.Contains(descLower, "algorithm not enforced") {
		return `// Specify the algorithm in the JWT verification
const decoded = jwt.verify(token, config.jwtSecret, { 
  algorithms: ['HS256'] // Explicitly specify the algorithms to use
});`
	} else if strings.Contains(descLower, "case-sensitive comparison") {
		return `// Use case-insensitive comparison for email
if (config.adminUsers.some(email => email.toLowerCase() === req.user.email.toLowerCase())) {
  next();
} else {
  res.status(403).json({ message: "Admin access required" });
}`
	} else if strings.Contains(descLower, "input sanitization") {
		return `// Use a comprehensive sanitization library
const sanitizeHtml = require('sanitize-html');

function sanitizeInput(input) {
  if (typeof input === 'string') {
    return sanitizeHtml(input, {
      allowedTags: ['b', 'i', 'em', 'strong', 'a'],
      allowedAttributes: {
        'a': ['href', 'target']
      }
    });
  }
  return input;
}`
	} else if strings.Contains(descLower, "timing attack") {
		return `// Use a constant-time comparison function
const crypto = require('crypto');

function constantTimeCompare(a, b) {
  if (typeof a !== 'string' || typeof b !== 'string') {
    return false;
  }
  
  // Create buffers of same length
  const aBuffer = Buffer.from(a);
  const bBuffer = Buffer.from(b);
  
  // Use crypto's timingSafeEqual
  try {
    return crypto.timingSafeEqual(aBuffer, bBuffer);
  } catch (e) {
    return false;
  }
}`
	} else if strings.Contains(descLower, "jwt") && strings.Contains(descLower, "expiration") {
		return `// Use stronger secret and reasonable expiration
const token = jwt.sign(
  user, 
  process.env.JWT_SECRET || crypto.randomBytes(32).toString('hex'), 
  { 
    expiresIn: '1h' // Shorter expiration time
  }
);`
	} else if strings.Contains(descLower, "idor") {
		return `// Verify the user ID properly
app.get("/api/user/profile", authMiddleware, async (req, res) => {
  try {
    // Get the requested userId
    const requestedUserId = req.query.id;
    
    // If another user's ID is requested, check if admin
    if (requestedUserId && requestedUserId !== req.user.id) {
      if (!req.user.isAdmin) {
        return res.status(403).json({ 
          message: "Access denied: You can only access your own profile" 
        });
      }
    }
    
    // Use authenticated user's ID or verified requested ID
    const userId = requestedUserId || req.user.id;
    
    // ... rest of the function
  } catch (error) {
    // ...
  }
});`
	} else if strings.Contains(descLower, "ssrf") {
		return `// Implement URL validation to prevent SSRF
const URL = require('url').URL;

app.post("/api/data/import", authMiddleware, async (req, res) => {
  try {
    const { url } = req.body;

    if (!url) {
      return res.status(400).json({ message: "URL required" });
    }

    // Validate URL and check against allowlist
    try {
      const parsedUrl = new URL(url);
      
      // Define allowlist of domains
      const allowedDomains = ['api.example.com', 'data.example.org'];
      
      if (!allowedDomains.includes(parsedUrl.hostname)) {
        return res.status(403).json({ 
          message: "Domain not allowed for security reasons" 
        });
      }
      
      // Ensure protocol is https
      if (parsedUrl.protocol !== 'https:') {
        return res.status(403).json({ 
          message: "Only HTTPS URLs are allowed" 
        });
      }
      
      // Now safe to make the request
      const response = await axios.get(url);
      res.json({ message: "Import successful", count: response.data.length });
    } catch (urlError) {
      return res.status(400).json({ message: "Invalid URL" });
    }
  } catch (error) {
    // ... error handling
  }
});`
	} else if strings.Contains(descLower, "prototype pollution") {
		return `// Prevent prototype pollution
const safeObjectAssign = (target, source) => {
  // Create a new object to avoid direct prototype pollution
  const result = Object.assign({}, target);
  
  // Only copy own properties, ignoring prototype properties
  if (source && typeof source === 'object') {
    Object.keys(source).forEach(key => {
      // Prevent __proto__ or constructor assignment
      if (key !== '__proto__' && key !== 'constructor' && 
          Object.prototype.hasOwnProperty.call(source, key)) {
        result[key] = source[key];
      }
    });
  }
  
  return result;
};

// Use the safe function
const mergedOptions = safeObjectAssign(defaultOptions, options);`
	} else if strings.Contains(descLower, "information disclosure") && !strings.Contains(descLower, "error logging") {
		return `// Remove sensitive information from system info
app.get("/api/system/info", authMiddleware, isAdmin, (req, res) => {
  try {
    // Only expose non-sensitive information
    const sysInfo = {
      environment: process.env.NODE_ENV,
      nodeVersion: process.versions.node,
      uptime: process.uptime(),
      // Do not include environment variables or file paths
    };

    res.json(sysInfo);
  } catch (error) {
    // ... error handling
  }
});`
	} else if strings.Contains(descLower, "encryption parameters") {
		return `// Improve encryption with stronger parameters and random IV
const encryptionService = {
  encrypt: (data, userKey) => {
    try {
      // Use stronger key derivation with more iterations and better hash
      const salt = crypto.randomBytes(16);
      const key = crypto.pbkdf2Sync(
        userKey || process.env.ENCRYPTION_KEY,
        salt,
        100000, // Increased iterations
        32,
        'sha256' // Stronger hash algorithm
      );

      // Random IV for each encryption
      const iv = crypto.randomBytes(16);
      const cipher = crypto.createCipheriv('aes-256-gcm', key, iv);
      
      let encrypted = cipher.update(JSON.stringify(data), 'utf8', 'hex');
      encrypted += cipher.final('hex');
      
      // Include auth tag for GCM mode
      const authTag = cipher.getAuthTag();
      
      // Return all parameters needed for decryption
      return {
        encrypted,
        iv: iv.toString('hex'),
        salt: salt.toString('hex'),
        authTag: authTag.toString('hex')
      };
    } catch (error) {
      // ... error handling
    }
  }
};`
	} else if strings.Contains(descLower, "template injection") {
		return `// Use a safe templating library
const handlebars = require('handlebars');

function renderTemplate(template, data) {
  try {
    // Compile the template using Handlebars
    const compiledTemplate = handlebars.compile(template);
    
    // Render the template with the provided data
    return compiledTemplate(data);
  } catch (error) {
    console.error('Template rendering error:', error.message);
    return '';
  }
}`
	}
	
	// Default case if no specific match
	return "";
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