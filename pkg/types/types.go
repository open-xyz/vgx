package types

// Vulnerability represents a security issue found in code
type Vulnerability struct {
	File        string `json:"file"`
	Description string `json:"description"`
	Rule        string `json:"rule,omitempty"`
	Severity    string `json:"severity,omitempty"`
	Line        int    `json:"line,omitempty"`
	Source      string `json:"source,omitempty"` // "semgrep" or "openai"
} 