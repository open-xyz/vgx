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
