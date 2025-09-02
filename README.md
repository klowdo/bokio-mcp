# 🧾 Bokio MCP Server

> Because your AI assistant shouldn't need an accounting degree to manage your books

A Model Context Protocol (MCP) server that brings the power of [Bokio](https://www.bokio.se) accounting to AI assistants. Built with Go, OAuth2, and a sprinkle of Nordic efficiency.

## ✨ Features

### 🔐 **Secure OAuth2 Authentication**
Full OAuth2 flow implementation with token management and automatic refresh

### 📊 **Complete Bokio API Coverage**
- **Invoices** - Create, list, update, and manage sales invoices
- **Customers** - Full CRUD operations for customer management
- **Journal Entries** - Accounting journal operations with reversal support
- **File Uploads** - Document and attachment management
- **Authentication** - Secure connection management and status checking

### 🛡️ **Read-Only Mode**
Enable `BOKIO_READ_ONLY=true` to prevent all write operations while maintaining full read access - perfect for AI assistants that should observe but not modify.

### 🚀 **Production Ready**
- Structured logging with slog
- Graceful shutdown handling
- Rate limiting and retry logic
- Comprehensive error handling
- Type-safe API with generated OpenAPI types

## 🎯 Quick Start

### Installation Options

#### Using Nix (Recommended)
```bash
# Enter development environment
nix develop

# Or with direnv
direnv allow
```

#### Using Docker
```bash
docker run -it \
  -e BOKIO_CLIENT_ID=your_client_id \
  -e BOKIO_CLIENT_SECRET=your_client_secret \
  ghcr.io/klowdo/bokio-mcp:latest
```

#### From Source
```bash
# Clone the repository
git clone https://github.com/klowdo/bokio-mcp.git
cd bokio-mcp

# Build the server
make build

# Run the server
./bin/bokio-mcp
```

## 🔧 Configuration

Configure the server using environment variables:

```bash
# Required - OAuth2 credentials
export BOKIO_CLIENT_ID="your_client_id"
export BOKIO_CLIENT_SECRET="your_client_secret"

# Optional - API configuration
export BOKIO_BASE_URL="https://api.bokio.se"      # Default
export BOKIO_REDIRECT_URL="http://localhost:8080/callback"  # Default

# Optional - Security
export BOKIO_READ_ONLY="true"  # Enable read-only mode
```

### Example `.env` file
```env
BOKIO_CLIENT_ID=your_client_id_here
BOKIO_CLIENT_SECRET=your_client_secret_here
BOKIO_READ_ONLY=false
```

## 📚 Available MCP Tools

### Authentication Tools
- `bokio_authenticate` - Start OAuth2 authentication flow
- `bokio_exchange_token` - Exchange authorization code for access token
- `bokio_get_connections` - Get current connection information
- `bokio_check_auth` - Check authentication status

### Invoice Tools
- `bokio_list_invoices` - List invoices with filtering and pagination
- `bokio_get_invoice` - Get specific invoice details
- `bokio_create_invoice` - Create new sales invoice
- `bokio_update_invoice` - Update existing invoice

### Customer Tools
- `bokio_list_customers` - List customers with pagination
- `bokio_get_customer` - Get specific customer details
- `bokio_create_customer` - Create new customer
- `bokio_update_customer` - Update customer information

### Journal Tools
- `bokio_list_journal_entries` - List journal entries
- `bokio_create_journal_entry` - Create new journal entry
- `bokio_reverse_journal_entry` - Reverse an existing entry
- `bokio_get_journal_entry` - Get specific journal entry

### Upload Tools
- `bokio_upload_file` - Upload documents and attachments
- `bokio_list_uploads` - List uploaded files
- `bokio_get_upload` - Get upload metadata
- `bokio_download_file` - Download uploaded file
- `bokio_delete_upload` - Delete uploaded file

## 🛠️ Development

### Prerequisites
- Go 1.23+
- Make
- Git

### Development Commands
```bash
# Update OpenAPI schemas
make update-schema

# Generate types from schemas
make generate-types

# Run tests
make test

# Run linting
make lint

# Run development server with hot reload
make dev

# Run security scans
make security

# Clean build artifacts
make clean

# See all available commands
make help
```

### Project Structure
```
bokio-mcp/
├── main.go              # Entry point and server setup
├── bokio/
│   ├── client.go        # Bokio API client with OAuth2
│   └── types.go         # API type definitions
├── tools/
│   ├── auth.go          # Authentication tools
│   ├── invoices.go      # Invoice management tools
│   ├── customers.go     # Customer management tools
│   ├── journal.go       # Journal entry tools
│   └── uploads.go       # File upload tools
├── schemas/             # OpenAPI specifications
├── Makefile            # Development automation
└── flake.nix           # Nix development environment
```

## 🧪 Testing

The server includes comprehensive test coverage:

```bash
# Run all tests
make test

# Run with race detection
go test -race ./...

# Generate coverage report
make test
open coverage.html
```

## 🚢 Deployment

### Docker Deployment
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bokio-mcp .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/bokio-mcp /bokio-mcp
ENTRYPOINT ["/bokio-mcp"]
```

### Systemd Service
```ini
[Unit]
Description=Bokio MCP Server
After=network.target

[Service]
Type=simple
User=bokio
ExecStart=/usr/local/bin/bokio-mcp
Restart=always
EnvironmentFile=/etc/bokio-mcp/env

[Install]
WantedBy=multi-user.target
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Bokio](https://www.bokio.se) for their excellent accounting API
- [Anthropic](https://anthropic.com) for the Model Context Protocol specification
- The Go community for excellent OAuth2 and HTTP libraries

## 🔗 Links

- [Bokio API Documentation](https://developer.bokio.se)
- [Model Context Protocol](https://modelcontextprotocol.io)
- [Project Issues](https://github.com/klowdo/bokio-mcp/issues)

---

Built with ❤️ and ☕ by [klowdo](https://github.com/klowdo)