# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bokio-mcp .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 bokio && \
    adduser -D -u 1000 -G bokio bokio

# Copy binary from builder
COPY --from=builder /app/bokio-mcp /usr/local/bin/bokio-mcp

# Set ownership
RUN chown bokio:bokio /usr/local/bin/bokio-mcp

# Switch to non-root user
USER bokio

# Set working directory
WORKDIR /home/bokio

# Expose port if needed (MCP uses stdio by default)
# EXPOSE 8080

# Run the MCP server
ENTRYPOINT ["bokio-mcp"]
