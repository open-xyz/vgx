# VibePenTester Integration Demo Guide

This guide provides step-by-step instructions for demonstrating the VGX security scanner with VibePenTester integration.

## Setup

1. Configure your environment:

```bash
# Copy the environment example file
cp .env.example .env

# Edit the .env file and add your API keys
# Make sure VIBE_ENABLED=true
```

2. Build and install the tools:

```bash
# Build the tools
make build

# OR install them globally
make install
```

## Demo 1: Basic Scanning with VGX

First, demonstrate the basic VGX scanning capability:

```bash
# Scan the JavaScript vulnerable example
vgx examples/vulnerable.js
```

Notice how VGX identifies vulnerabilities using its built-in scanners.

## Demo 2: Contextual Scanning

Now show how VGX can scan files with codebase context awareness:

```bash
# First scan to build the context
vgx --update-context=true examples/vulnerable.js

# Make a change to the file
# For example, add another vulnerability or modify an existing one

# Scan just the changed files with context
vgx --changes=true examples/vulnerable.js

# Show how the recommendations are more precise due to context
```

Highlight how the scanner provides better, more contextual recommendations that understand the codebase structure.

## Demo 3: VibePenTester CLI Tool

Now demonstrate the dedicated VibePenTester CLI tool:

```bash
# Scan the vulnerable files with the vibe tool
vibe scan examples/vulnerable.js examples/vulnerable.py

# Scan with text output
vibe scan -format text examples/vulnerable.js

# Scan recursively through a directory
vibe scan -recursive examples/
```

## Demo 4: Generate Security Reports

Show how to generate professional security reports:

```bash
# First run a scan and save results to JSON
vibe scan -output scan-results.json examples/

# Generate an HTML report
vibe report -input scan-results.json -output security-report.html

# Generate a Markdown report
vibe report -input scan-results.json -format markdown -output security-report.md
```

Open the generated HTML report in a browser to showcase the report interface.

## Demo 5: Integration Highlights

During the presentation, highlight these key points:

1. **Multiple Scanner Integration**: VGX combines rule-based scanning (Semgrep), AI analysis (OpenAI), and VibePenTester's comprehensive security testing.

2. **Configurable Options**: Show the various environment variables in `.env` that allow customization.

3. **Pre-Commit Hook**: Demonstrate how VGX can be set up as a pre-commit hook to catch vulnerabilities before they enter the codebase.

4. **Report Generation**: Emphasize the detailed, professional reports for security audits and team sharing.

## Example Vulnerabilities to Highlight

Use these specific examples from the vulnerable files to demonstrate scanning capabilities:

### JavaScript (`examples/vulnerable.js`):

- SQL Injection in user query
- Command Injection in the ping endpoint
- Hardcoded API credentials
- Cross-site scripting vulnerability in search

### Python (`examples/vulnerable.py`):

- SQL Injection in user lookup
- Insecure deserialization with pickle
- Path traversal vulnerability
- Information disclosure through error messages

## Fixing Issues

As an optional part of the demo, demonstrate fixing some of the vulnerabilities:

1. SQL Injection fix - Show parameterized queries
2. XSS fix - Show proper output escaping
3. Credentials fix - Show using environment variables

## Conclusion

Wrap up by emphasizing the security advantages of this integrated approach:

- Early vulnerability detection
- Multiple detection methods for better coverage
- Professional reporting
- Seamless integration into development workflow
