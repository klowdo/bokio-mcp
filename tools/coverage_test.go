package tools

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/klowdo/bokio-mcp/bokio/generated/general"
	"github.com/stretchr/testify/require"
)

var uncoveredByDesign = map[string]string{
	"Authorize":                    "OAuth2 authorization redirect, unusable with integration token auth",
	"RequestToken":                 "OAuth2 token exchange, unusable with integration token auth",
	"RequestTokenWithFormdataBody": "OAuth2 token exchange, unusable with integration token auth",
	"RequestTokenWithBody":         "OAuth2 token exchange, unusable with integration token auth",
}

func toolSources(t *testing.T) string {
	t.Helper()

	paths, err := filepath.Glob("*.go")
	require.NoError(t, err)

	var sources strings.Builder
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		sources.Write(content)
	}
	return sources.String()
}

func TestEveryGeneratedEndpointIsExposed(t *testing.T) {
	sources := toolSources(t)

	isCalled := func(method string) bool {
		base := strings.TrimSuffix(method, "WithBody")
		return strings.Contains(sources, "."+base+"(") || strings.Contains(sources, "."+base+"WithBody(")
	}

	var missing []string
	for _, clientType := range []reflect.Type{
		reflect.TypeOf(&company.Client{}),
		reflect.TypeOf(&general.Client{}),
	} {
		for i := range clientType.NumMethod() {
			method := clientType.Method(i).Name
			if _, skipped := uncoveredByDesign[method]; skipped || isCalled(method) {
				continue
			}
			missing = append(missing, method)
		}
	}

	require.Emptyf(t, missing, "generated endpoints with no MCP tool: %s", strings.Join(missing, ", "))
}
