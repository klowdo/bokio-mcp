// Package generated contains auto-generated API clients and types for the Bokio API.
//
// Generation is handled via go generate directives below.
// Run: go generate ./bokio/generated/...
//
//go:generate sh -c "go tool oapi-codegen -package company -generate types ../../schemas/company-api.yaml > company/types.go"
//go:generate sh -c "go tool oapi-codegen -package company -generate client ../../schemas/company-api.yaml > company/client.go"
//go:generate sh -c "go tool oapi-codegen -package general -generate types ../../schemas/general-api.yaml > general/types.go"
//go:generate sh -c "go tool oapi-codegen -package general -generate client ../../schemas/general-api.yaml > general/client.go"
package generated
