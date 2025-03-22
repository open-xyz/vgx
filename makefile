# Makefile for VibeGuard (VGX)

# Build the CLI binary
build:
	go build -o vgx ./cmd/vgx

# Install the CLI binary
install:
	go install ./cmd/vgx

# Build and push Docker image
docker-build:
	docker build -t vgx .

# Run tests
test:
	go test ./...

# Clean up generated files
clean:
	rm -f vgx
	docker rmi vgx
