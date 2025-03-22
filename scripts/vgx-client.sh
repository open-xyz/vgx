#!/bin/bash
# scripts/vgx-client.sh

set -e

# Check if VGX service is running
if ! docker ps | grep -q vgx-service; then
    echo "❌ VGX service is not running. Starting it now..."
    docker start vgx-service
fi

# Handle install-hook command
if [ "$1" = "install-hook" ]; then
    mkdir -p .git/hooks
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# VGX pre-commit hook
vgx scan
EOF
    chmod +x .git/hooks/pre-commit
    echo "✅ Pre-commit hook installed successfully!"
    exit 0
fi

# Handle scan command
if [ "$1" = "scan" ] || [ -z "$1" ]; then
    # Get list of staged files
    files=$(git diff --cached --name-only --diff-filter=ACMR)
    
    if [ -z "$files" ]; then
        echo "No files staged for commit."
        exit 0
    fi
    
    # Call the VGX service API
    response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"files\": [$(echo $files | sed 's/ /","/g' | sed 's/^/"/' | sed 's/$/"/')], \"repo\": \"$(git rev-parse --show-toplevel)\"}" \
        http://localhost:9977/scan)
    
    # Parse the response
    if echo "$response" | grep -q "\"status\":\"error\""; then
        echo "❌ Error scanning files:"
        echo "$response" | jq -r '.message'
        exit 1
    fi
    
    vulnerabilities=$(echo "$response" | jq -r '.vulnerabilities | length')
    
    if [ "$vulnerabilities" -gt 0 ]; then
        echo "🚨 VGX found $vulnerabilities vulnerabilities:"
        echo "$response" | jq -r '.vulnerabilities[] | "• " + .file + ": " + .description'
        exit 1
    else
        echo "✅ VGX: No vulnerabilities found!"
        exit 0
    fi
fi

# Help message for other commands
echo "VGX - Security scanner"
echo "Usage:"
echo "  vgx scan                  - Scan staged files"
echo "  vgx install-hook          - Install pre-commit hook"
echo "  vgx update                - Update VGX to latest version"
