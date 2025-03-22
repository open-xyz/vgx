#!/bin/bash
# scripts/install.sh

set -e

# Check for Docker
if ! command -v docker &> /dev/null; then
    echo "Docker is required but not installed. Please install Docker first."
    exit 1
fi

# Pull the image
echo "📥 Pulling VGX image..."
docker pull ghcr.io/yourusername/vgx:latest

# Create configuration directory
mkdir -p ~/.vgx/cache

# Ask for OpenAI API key if not set
if [ -z "$OPENAI_API_KEY" ]; then
    echo -n "Enter your OpenAI API key (or set OPENAI_API_KEY environment variable): "
    read -r api_key
    echo "OPENAI_API_KEY=$api_key" > ~/.vgx/.env
    echo "API key saved to ~/.vgx/.env"
fi

# Start the service
echo "🚀 Starting VGX service..."
docker run -d \
    --name vgx-service \
    --restart unless-stopped \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v ~/.vgx/cache:/app/cache \
    -p 127.0.0.1:9977:8080 \
    --env-file ~/.vgx/.env \
    ghcr.io/yourusername/vgx:latest

# Install the client script
echo "📝 Installing VGX client..."
sudo curl -L -o /usr/local/bin/vgx https://raw.githubusercontent.com/yourusername/vgx/main/scripts/vgx-client.sh
sudo chmod +x /usr/local/bin/vgx

echo "✅ VGX installed successfully!"
echo "Run 'vgx install-hook' in any Git repository to set up the pre-commit hook."
