FROM golang:1.18-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the applications
RUN make build

FROM alpine:3.16

# Install runtime dependencies
RUN apk add --no-cache git python3 py3-pip bash

# Install semgrep
RUN pip3 install semgrep

# Set working directory
WORKDIR /scan

# Copy the built binaries from the builder stage
COPY --from=builder /app/vgx /usr/local/bin/vgx
COPY --from=builder /app/vibe /usr/local/bin/vibe

# Copy the example .env file
COPY .env.example /etc/vgx/.env.example

# Create volume for persistent context
VOLUME ["/root/.vgx"]

# Set entrypoint for the container
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["docker-entrypoint.sh"]

# Default command is to show usage
CMD ["--help"] 