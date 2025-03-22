FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o vgx-service ./cmd/service

FROM alpine:latest
# Install git and Python
RUN apk add --no-cache git python3 py3-pip

# Create a virtual environment and install semgrep
RUN python3 -m venv /opt/venv && \
    /opt/venv/bin/pip install semgrep

# Add virtual environment to PATH
ENV PATH="/opt/venv/bin:${PATH}"

# Copy the built service binary
COPY --from=builder /app/vgx-service /usr/local/bin/

# Copy client script
COPY scripts/vgx-client.sh /usr/local/bin/vgx

# Make the script executable
RUN chmod +x /usr/local/bin/vgx

EXPOSE 8080
VOLUME ["/code"]
ENTRYPOINT ["vgx-service"]
