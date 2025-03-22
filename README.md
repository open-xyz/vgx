# VGX

A Git pre-commit security scanner with OpenAI integration to detect vulnerabilities before they enter your codebase.

## Features

- 🔍 **Pre-commit scanning**: Automatically scan staged files before each commit
- 🤖 **AI-powered analysis**: Leverage OpenAI to detect complex security vulnerabilities
- 🛡️ **Semgrep integration**: Use rule-based scanning alongside AI detection
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

## Configuration

- Create a `.env` file in your project root:

```bash
OPENAI_API_KEY=your-api-key-here
```

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

    Fork the repository
    Create your feature branch (git checkout -b feature/amazing-feature)
    Commit your changes (git commit -m 'Add some amazing feature')
    Push to the branch (git push origin feature/amazing-feature)
    Open a Pull Request
```

## License

- Distributed under the MIT License. See `LICENSE` for more information.
