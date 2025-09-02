//go:build ignore

// Package generated contains auto-generated API clients and types for the Bokio API.
// 
// Generation is handled via go generate directives below.
// Run: go generate ./bokio/generated/...
//
//go:generate sh -c "go tool oapi-codegen -package company -generate types ../../schemas/company-api.yaml > company/types.go"
//go:generate sh -c "go tool oapi-codegen -package company -generate client,skip-fmt ../../schemas/company-api.yaml > company/client.go"
//go:generate sh -c "go tool oapi-codegen -package general -generate types ../../schemas/general-api.yaml > general/types.go"  
//go:generate sh -c "go tool oapi-codegen -package general -generate client,skip-fmt ../../schemas/general-api.yaml > general/client.go"
//go:generate goimports -w company/ general/
//go:generate gofmt -w company/ general/
package generated