package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestToolSchemasAreObjects(t *testing.T) {
	client, err := bokio.NewAuthClient(&bokio.Config{
		IntegrationToken: "test-token",
		BaseURL:          "https://api.bokio.se",
	})
	require.NoError(t, err)

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	require.NoError(t, RegisterGeneratedJournalTools(server, client))
	require.NoError(t, RegisterCustomerTools(server, client))
	require.NoError(t, RegisterItemTools(server, client))
	require.NoError(t, RegisterInvoiceTools(server, client))
	require.NoError(t, RegisterUploadTools(server, client))

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	clientSession, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).
		Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer clientSession.Close()

	listed, err := clientSession.ListTools(ctx, nil)
	require.NoError(t, err)

	count := 0
	for _, tool := range listed.Tools {
		count++
		for name, schema := range map[string]any{"input": tool.InputSchema, "output": tool.OutputSchema} {
			if schema == nil {
				continue
			}
			raw, err := json.Marshal(schema)
			require.NoError(t, err)
			assertObjectSchema(t, tool.Name+"."+name, raw)
		}
	}
	require.NotZero(t, count)
}

func assertObjectSchema(t *testing.T, path string, raw []byte) {
	t.Helper()

	var node map[string]json.RawMessage
	require.NoErrorf(t, json.Unmarshal(raw, &node),
		"%s: schema must be an object, not %s (MCP clients reject boolean schemas)", path, raw)

	_, typed := node["type"]
	_, multiTyped := node["types"]
	require.Truef(t, typed || multiTyped,
		"%s: schema %s has no type (an interface{} field infers to {}, and MCP clients reject it)", path, raw)

	var props map[string]json.RawMessage
	if err := json.Unmarshal(node["properties"], &props); err != nil {
		return
	}
	for name, prop := range props {
		assertObjectSchema(t, path+"."+name, prop)
	}
}
