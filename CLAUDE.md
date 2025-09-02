# CLAUDE.md - Bokio MCP Server Development Guide

This file provides guidance to AI assistants (particularly Claude Code) when working on the Bokio MCP (Model Context Protocol) server project.

## 📋 Project Overview

The Bokio MCP Server is a Go-based Model Context Protocol server that provides AI assistants with secure access to the Bokio accounting API. It implements OAuth2 authentication, comprehensive API coverage, and includes a read-only mode for safe AI interactions.

### Key Technologies
- **Go 1.24** with tool directives (uses `go tool` commands, NOT `go run`)
- **Model Context Protocol (MCP)** for AI assistant integration
- **OAuth2** authentication with token management  
- **OpenAPI code generation** using oapi-codegen
- **Nix flake** for reproducible development environment
- **direnv** for automatic environment loading

## 🏗️ Project Architecture

### Directory Structure
```
bokio-mcp/
├── main.go                 # Entry point, server setup, tool registration
├── bokio/                  # Bokio API client and types
│   ├── client.go           # OAuth2 client implementation
│   ├── types.go            # Manual type definitions
│   └── generated/          # Generated types and clients from OpenAPI
├── tools/                  # MCP tool implementations
│   ├── auth.go            # Authentication tools
│   ├── invoices.go        # Invoice management tools
│   ├── customers.go       # Customer management tools
│   ├── journal.go         # Journal entry tools
│   └── uploads.go         # File upload tools
├── schemas/               # OpenAPI specifications (downloaded)
├── Makefile              # 25+ development automation targets
├── flake.nix             # Nix development environment
├── go.mod                # Go 1.24 with tool directives
└── .envrc                # direnv configuration
```

### Core Components

#### MCP Tool Registration Pattern
All MCP tools are registered in `main.go` using this pattern:
```go
// Register tools with the server
if err := tools.RegisterAuthTools(server, bokioClient); err != nil {
    return fmt.Errorf("failed to register auth tools: %w", err)
}
```

#### Read-Only Mode
The server supports a read-only mode via `BOKIO_READ_ONLY=true` environment variable that disables all write operations while maintaining full read access.

## 🔧 Development Workflow

### Environment Setup
```bash
# Option 1: Using Nix (recommended)
nix develop

# Option 2: Using direnv (auto-loads when entering directory)
direnv allow

# Verify setup
make info  # Shows project status and tool availability
```

### Daily Development Commands
```bash
# Update API schemas and regenerate types
make update-schema && make generate-types

# Build the project
make build

# Run development server
make dev

# Run tests with coverage
make test

# Code quality checks
make lint

# Security scanning
make security

# Full pre-commit pipeline
make pre-commit
```

### Go 1.24 Tool Directives
**CRITICAL**: This project uses Go 1.24's tool directive system. Always use `go tool` commands, never `go run`:

```bash
# ✅ Correct (automated via Makefile)
go tool oapi-codegen -package generated -generate types schema.yaml
go tool golangci-lint run ./...
go tool gosec ./...

# ❌ Wrong - DO NOT USE
go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
```

Tools are declared in `go.mod`:
```go
tool (
    github.com/deepmap/oapi-codegen/cmd/oapi-codegen
    github.com/golangci/golangci-lint/cmd/golangci-lint
    github.com/securego/gosec/v2/cmd/gosec
    golang.org/x/vuln/cmd/govulncheck
)
```

## 📝 Commit Message Conventions

Follow these conventions for consistent git history:

### Bug Fixes
```
fix: resolve OAuth token refresh issue
fix: handle empty customer list response
fix: validate invoice date format in create tool
```

### New Features
```
feat: add journal entry reversal support
feat: implement file upload tool with progress tracking
feat: add pagination support to customer listing
```

### API Schema Updates
```
schema: update Bokio API specifications to latest version
schema: add new invoice status fields from API v2.1
```

### Generated Code Updates
```
gen: regenerate types from updated API schemas
gen: update client code after schema changes
```

### Build System Changes
```
build: update Go to 1.24 and migrate to tool directives
build: add security scanning to CI pipeline
build: optimize Docker build with multi-stage approach
```

### Documentation
```
docs: add MCP usage examples to README
docs: update API authentication flow documentation
```

### Refactoring
```
refactor: extract common OAuth error handling
refactor: simplify MCP tool registration pattern
```

## 🔄 API Schema Update Procedures

### Complete Schema Update Workflow
```bash
# 1. Download latest schemas from Bokio GitHub
make update-schema

# 2. Regenerate Go types and clients
make generate-types

# 3. Verify build still works
make build

# 4. Run tests to catch breaking changes
make test

# 5. Check for linting issues
make lint

# 6. Test authentication flows
make dev  # Test OAuth flow manually

# 7. Commit changes in logical steps
git add schemas/
git commit -m "schema: update Bokio API specifications to latest version"

git add bokio/generated/
git commit -m "gen: regenerate types from updated API schemas"

# 8. If API changes require code updates, commit separately
git add tools/ bokio/client.go
git commit -m "feat: adapt to new API endpoints in schema update"
```

### Schema Sources
- **Company API**: `https://raw.githubusercontent.com/bokio/bokio-api/v1/api-specification/company-api.yaml`
- **General API**: `https://raw.githubusercontent.com/bokio/bokio-api/v1/api-specification/general-api.yaml`

### Generated Files (Never Edit Manually)
- `bokio/generated/company_types.go`
- `bokio/generated/general_types.go` 
- `bokio/generated/company_client.go`
- `bokio/generated/general_client.go`

## 🧪 Testing & Quality Assurance

### Testing Strategy
```bash
# Run all tests with race detection and coverage
make test

# View coverage report in browser
open coverage.html

# Run specific test packages
go test ./tools/... -v
```

### Code Quality Pipeline
```bash
# Lint with golangci-lint (comprehensive rules)
make lint

# Security scanning
make security  # Runs govulncheck and gosec

# Format code
make format

# Full pre-commit pipeline
make pre-commit  # deps → lint → test → security → build
```

### Development Server Testing
```bash
# Start development server with enhanced logging
make dev

# In another terminal, test MCP tools
# (use your MCP client to test authentication flow)
```

## 🔐 Authentication & Security

### OAuth2 Flow Implementation
1. **Start Authentication**: Use `bokio_authenticate` tool
2. **User Authorization**: User visits returned URL and authorizes
3. **Token Exchange**: Use `bokio_exchange_token` with authorization code
4. **Connection Verification**: Use `bokio_check_auth` to verify status

### Read-Only Mode
Enable for safe AI assistant operations:
```bash
export BOKIO_READ_ONLY=true
```

This disables all write operations (`create_*`, `update_*`, `delete_*`) while maintaining full read access.

## 🛠️ MCP Tool Development

### Adding New MCP Tools

1. **Create Tool File** in `tools/` directory
2. **Define Parameter & Result Structs**:
```go
type NewToolParams struct {
    Field string `json:"field"`
}

type NewToolResult struct {
    Success bool   `json:"success"`
    Data    string `json:"data,omitempty"`
    Error   string `json:"error,omitempty"`
}
```

3. **Implement Tool Registration Function**:
```go
func RegisterNewTools(server *mcp.Server, client *bokio.Client) error {
    tool := mcp.NewServerTool[NewToolParams, NewToolResult](
        "bokio_new_action",
        "Description of what this tool does",
        func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[NewToolParams]) (*mcp.CallToolResultFor[NewToolResult], error) {
            // Implementation here
        },
        mcp.Input(
            mcp.Property("field", 
                mcp.Description("Field description"),
                mcp.Required(true),
            ),
        ),
    )
    
    server.AddTools(tool)
    return nil
}
```

4. **Register in main.go**:
```go
if err := tools.RegisterNewTools(server, bokioClient); err != nil {
    return fmt.Errorf("failed to register new tools: %w", err)
}
```

### Read-Only Mode Implementation
Check for read-only mode in write operations:
```go
if client.GetConfig().ReadOnly {
    return &mcp.CallToolResultFor[Result]{
        Content: []mcp.Content{
            &mcp.TextContent{
                Text: "Operation not allowed in read-only mode",
            },
        },
    }, nil
}
```

## 🚀 Release Management

### Pre-Release Checklist
```bash
# 1. Update schemas and regenerate types
make update-schema && make generate-types

# 2. Run full test suite
make pre-commit

# 3. Test release build
make release-dry

# 4. Verify Nix build works
make nix-build

# 5. Update version and create tag
make tag VERSION=v1.x.x

# 6. Create release
make release
```

### Version Management
- Use semantic versioning (v1.2.3)
- Tag releases with `make tag VERSION=vX.Y.Z`
- Use `make release-dry` to test GoReleaser configuration

## 🔍 Troubleshooting

### Common Issues

#### "go tool: command not found"
- **Cause**: Using wrong Go version or tools not installed
- **Solution**: Ensure Go 1.24 and use `nix develop` environment

#### Generated files out of sync
- **Cause**: Schema updated but types not regenerated  
- **Solution**: `make clean && make generate-types`

#### OAuth authentication fails
- **Cause**: Invalid client credentials or redirect URI mismatch
- **Solution**: Verify `BOKIO_CLIENT_ID`, `BOKIO_CLIENT_SECRET`, and `BOKIO_REDIRECT_URL`

#### MCP client connection issues
- **Cause**: Server not running or stdio transport issues
- **Solution**: Check server logs, verify MCP client configuration

### Debug Commands
```bash
# Show project status and dependencies
make info

# Check dependency health
make check-deps

# Verbose development server
make dev

# Build with debug information
go build -gcflags="all=-N -l" -o bin/bokio-mcp-debug .
```

## 📚 Useful Make Targets

The project includes 25+ Make targets for automation:

### Primary Development
- `make update-schema` - Download latest Bokio API specs
- `make generate-types` - Generate Go types from OpenAPI
- `make build` - Build the MCP server binary
- `make dev` - Run development server  
- `make test` - Run tests with coverage
- `make lint` - Code quality analysis
- `make security` - Security scanning

### Quality Assurance
- `make pre-commit` - Full pipeline (deps, lint, test, security, build)
- `make format` - Format all Go code
- `make clean` - Clean build artifacts

### Release Management
- `make release-dry` - Test release configuration
- `make tag VERSION=vX.Y.Z` - Create and push version tag
- `make release` - Create GitHub release

### Information
- `make info` - Show project status and configuration
- `make help` - Show all available targets

## 🎯 Development Best Practices

### Code Organization
- Keep MCP tools focused and single-purpose
- Use consistent error handling patterns
- Implement proper logging with slog
- Follow Go naming conventions

### API Integration
- Always validate API responses
- Implement proper rate limiting respect
- Handle authentication errors gracefully
- Use generated types for type safety

### Testing
- Write tests for all MCP tools
- Test both success and error scenarios
- Include authentication flow testing
- Test read-only mode restrictions

### Git Workflow
- Make atomic commits with clear messages
- Separate schema updates from code changes
- Test thoroughly before committing
- Use descriptive branch names

---

## 🤖 AI Assistant Guidelines

When working on this project:

1. **Always use `make` targets** instead of running commands directly
2. **Never edit generated files** in `bokio/generated/`
3. **Test authentication flows** after API changes
4. **Verify read-only mode** works for write operations
5. **Update schemas separately** from code changes
6. **Use Go 1.24 tool directives** - never `go run`
7. **Follow commit message conventions** for clear history
8. **Run `make pre-commit`** before final commits

Remember: This is a production MCP server handling financial data. Prioritize security, reliability, and clear error messages in all implementations.