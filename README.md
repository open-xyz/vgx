# VGX

A Git pre-commit security scanner with OpenAI and VibePenTester integration to detect vulnerabilities before they enter your codebase.

## Features

- 🔍 **Pre-commit scanning**: Automatically scan staged files before each commit
- 🤖 **AI-powered analysis**: Leverage OpenAI to detect complex security vulnerabilities
- 🛡️ **Semgrep integration**: Use rule-based scanning alongside AI detection
- 🔌 **VibePenTester integration**: Connect with VibePenTester for enhanced security analysis
- ⚡ **Fast & lightweight**: Written in Go for maximum performance
- 🔌 **Extensible**: Easy to customize and extend for your specific needs

## Installation

```bash
# Clone the repository
git clone https://github.com/open-xyz/vgx.git
cd vgx

# Install the CLI
make install

# Add Go bin to your PATH if needed
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

## Basic Usage

```bash
# Run a scan on staged files
vgx

# Specify files to scan
vgx file1.js file2.py
```

## VibePenTester CLI Tool

The VibePenTester integration includes a dedicated CLI tool for scanning files and generating reports:

```bash
# Scan files and display results
vibe scan file1.js file2.py

# Scan a directory (non-recursive)
vibe scan ./src

# Scan directories recursively
vibe scan -recursive ./src ./lib

# Save scan results to a JSON file
vibe scan -output results.json ./src

# Generate HTML report from scan results
vibe report -input results.json -output report.html

# Generate Markdown report
vibe report -input results.json -format markdown -output report.md
```

## Set Up Pre-commit Hook

```bash
# Navigate to your repository
cd /path/to/your/repo

# Install the pre-commit hook
mkdir -p .git/hooks
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/sh
vgx
EOF
chmod +x .git/hooks/pre-commit
```

## Bypassing the Hook (Emergency Override)

**Emergency Override:** To bypass the pre-commit hook in urgent situations:

```bash
git commit -m "Critical fix" --no-verify
```

## Configuration

- Create a `.env` file in your project root based on the example:

```bash
cp .env.example .env
# Edit the .env file with your API keys
```

### Available Configuration Options

| Environment Variable   | Description                         | Default                    |
| ---------------------- | ----------------------------------- | -------------------------- |
| `OPENAI_API_KEY`       | OpenAI API key for AI analysis      | Required for OpenAI        |
| `DISABLE_OPENAI`       | Set to `true` to disable OpenAI     | `false`                    |
| `VIBE_ENABLED`         | Enable VibePenTester integration    | `false`                    |
| `VIBE_API_KEY`         | API key for VibePenTester           | Required for VibePenTester |
| `VIBE_SERVER_URL`      | URL of VibePenTester service        | `http://localhost:5050`    |
| `VIBE_SCAN_SCOPE`      | Scan scope (url, domain, subdomain) | `url`                      |
| `VIBE_TIMEOUT_SECONDS` | Timeout for VibePenTester requests  | `60`                       |
| `VIBE_UPLOAD_LOGS`     | Upload scan logs to VibePenTester   | `false`                    |

## VibePenTester Integration

VGX integrates with [VibePenTester](https://github.com/yourusername/vibe_pen_tester) for enhanced security analysis:

1. Ensure VibePenTester is running locally or on a remote server
2. Configure the integration in your `.env` file:
   ```
   VIBE_ENABLED=true
   VIBE_API_KEY=your-vibepentester-api-key
   VIBE_SERVER_URL=http://your-vibepentester-server:5050
   ```
3. Run VGX as usual - it will now also include VibePenTester analysis results

This integration combines the strengths of rule-based scanning (Semgrep), AI analysis (OpenAI), and VibePenTester's comprehensive security testing capabilities.

## Development

```bash
# Build the CLI
make build

# Run tests
make test

# Build Docker image
make docker-build
```

## Contributing

```bash
# Fork the repository
# Create your feature branch (git checkout -b feature/amazing-feature)
# Commit your changes (git commit -m 'Add some amazing feature')
# Push to the branch (git push origin feature/amazing-feature)
# Open a Pull Request
```

## License

- Distributed under the MIT License. See `LICENSE` for more information.
