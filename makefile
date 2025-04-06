# Makefile for VibeGuard (VGX)

# Build all binaries
build: build-vgx build-vibe

# Build the VGX CLI binary
build-vgx:
	go build -o vgx ./cmd/vgx

# Build the VibePenTester CLI binary
build-vibe:
	go build -o vibe ./cmd/vibe

# Install all CLI tools
install: install-vgx install-vibe

# Install the VGX CLI binary
install-vgx:
	go install ./cmd/vgx

# Install the VibePenTester CLI binary
install-vibe:
	go install ./cmd/vibe

# Build and push Docker image
docker-build:
	docker build -t vgx .

# Run tests
test:
	go test ./...

# Clean up generated files
clean:
	rm -f vgx
	rm -f vibe
	docker rmi vgx
