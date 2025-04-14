# Scanning Any Codebase with VGX

This guide explains how to use VGX to scan any codebase for security vulnerabilities using Docker containers.

## Prerequisites

- Docker and Docker Compose installed
- Git installed (to detect code changes)
- (Optional) OpenAI API key for AI-powered scanning
- (Optional) VibePenTester API key for enhanced security analysis

## Quick Start

The easiest way to scan any codebase is to use the provided `scan-codebase.sh` script:

```bash
# Clone the VGX repository
git clone https://github.com/open-xyz/vgx.git
cd vgx

# Make the script executable
chmod +x scan-codebase.sh

# Scan a codebase (defaults to current directory)
./scan-codebase.sh /path/to/your/codebase

# Scan with OpenAI integration
./scan-codebase.sh --openai sk-your-openai-key /path/to/your/codebase

# Scan with VibePenTester
./scan-codebase.sh --vibe --vibe-key your-vibe-key /path/to/your/codebase

# Scan all files (not just changes)
./scan-codebase.sh --all /path/to/your/codebase

# Scan a specific file
./scan-codebase.sh --file src/vulnerable.js /path/to/your/codebase
```

## What Happens During a Scan

When you scan a codebase:

1. The scanner adds entries to `.gitignore` for VGX files (`reports/` and `.vgx/`)
2. It creates a `reports` directory in your codebase to store scan results
3. Results are stored in Markdown format in the `reports` directory
4. A context of your codebase is stored in a Docker volume for better change detection
5. Only changed files are scanned by default (unless `--all` is specified)

## Command Line Options

```
Usage: ./scan-codebase.sh [options] [path_to_codebase]

Options:
  -a, --all             Scan all files, not just changed ones
  -f, --file FILE       Scan a specific file or directory
  -v, --vibe            Enable VibePenTester integration
  -r, --report FORMAT   Generate report in specified format (html, md)
  -o, --openai KEY      Specify OpenAI API key
  -k, --vibe-key KEY    Specify VibePenTester API key
  -s, --server URL      Specify VibePenTester server URL
  -h, --help            Show this help message
```

## Using Docker Compose Directly

If you prefer to use Docker Compose manually:

```bash
# Set environment variables
export SCAN_TARGET=/path/to/your/codebase
export SCAN_CMD=scan-changes  # or "scan" for all files
export OPENAI_API_KEY=your-key-here  # optional
export VIBE_ENABLED=true  # optional
export VIBE_API_KEY=your-key-here  # optional

# Build and run
docker-compose build
docker-compose up
```

## Understanding the Results

After scanning, check the `reports` directory in your codebase for:

- Markdown files with details of found vulnerabilities
- Line numbers and descriptions of issues
- Recommendations for fixing the vulnerabilities
- Source of each vulnerability (OpenAI, VibePenTester, Semgrep)

## Persistent Context

The scanner maintains context about your codebase in a Docker volume (`vgx-context`). This allows it to:

1. Track changes between scans
2. Only scan files that have actually changed
3. Provide better recommendations by understanding the codebase structure

## Troubleshooting

If you encounter issues:

- Make sure Docker is running
- Check that your codebase is a Git repository (for change detection)
- Ensure you have proper permissions to write to the codebase directory
- For OpenAI scanning, verify your API key is valid
- For VibePenTester, ensure the service is running at the specified URL
