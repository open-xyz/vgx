#!/bin/bash
set -e

# Initialize environment
if [ ! -f .env ]; then
    echo "No .env file found in the mounted codebase, using default configuration"
    cp /etc/vgx/.env.example .env
fi

# Create reports directory in the codebase
mkdir -p reports

# Create .gitignore if it doesn't exist
if [ ! -f .gitignore ]; then
    echo "# VGX Security Scanner" > .gitignore
    echo "reports/" >> .gitignore
    echo ".vgx/" >> .gitignore
    echo ".env" >> .gitignore
    echo "Created .gitignore file with VGX entries"
else
    # Check if reports/ is already in .gitignore
    if ! grep -q "reports/" .gitignore; then
        echo "reports/" >> .gitignore
        echo "Added reports/ to .gitignore"
    fi
    
    # Check if .vgx/ is already in .gitignore
    if ! grep -q ".vgx/" .gitignore; then
        echo ".vgx/" >> .gitignore
        echo "Added .vgx/ to .gitignore"
    fi
fi

# Create local context directory (will be mounted from host for persistence)
mkdir -p /root/.vgx

# If the command is "scan", run our custom scanning logic
if [ "$1" = "scan" ]; then
    shift
    
    # Default to all files if no specific files are provided
    if [ $# -eq 0 ]; then
        echo "Scanning all files in the codebase..."
        vgx --changes=false --report=true --update-context=true
    else
        echo "Scanning specified files..."
        vgx --report=true --update-context=true "$@"
    fi
    
    echo "Scan complete! Check the 'reports' directory for detailed findings."
    exit 0
fi

# If the command is scan-changes, only scan changed files
if [ "$1" = "scan-changes" ]; then
    shift
    
    echo "Scanning changed files..."
    vgx --changes=true --report=true --update-context=true "$@"
    
    echo "Scan complete! Check the 'reports' directory for detailed findings."
    exit 0
fi

# If the command is vibe, run it directly
if [ "$1" = "vibe" ]; then
    shift
    exec vibe "$@"
    exit 0
fi

# For any other command, execute vgx with the provided arguments
exec vgx "$@" 