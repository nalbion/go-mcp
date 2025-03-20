package mcp_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/nalbion/go-mcp/pkg/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaValidation(t *testing.T) {
	// given
	schemaFile := "schema.json"

	// when we read the schema file
	schemaData, err := os.ReadFile(schemaFile)

	// then we can parse it successfully
	require.NoError(t, err, "Failed to read schema file")

	var schema map[string]interface{}
	err = json.Unmarshal(schemaData, &schema)
	require.NoError(t, err, "Failed to parse schema JSON")

	// Verify schema structure
	assert.Contains(t, schema, "components")
	assert.Contains(t, schema, "info")
	assert.Contains(t, schema, "openapi")
}

func TestToolModels(t *testing.T) {
	// given
	tool := mcp.Tool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]mcp.ToolInputSchemaProperty{
				"param1": {
					Type:        "string",
					Description: "A test parameter",
				},
			},
			Required: []string{"param1"},
		},
	}

	// when we marshal and unmarshal the tool
	data, err := json.Marshal(tool)
	require.NoError(t, err)

	var unmarshaled mcp.Tool
	err = json.Unmarshal(data, &unmarshaled)

	// then the tool is preserved correctly
	require.NoError(t, err)
	assert.Equal(t, "test-tool", unmarshaled.Name)
	assert.Equal(t, "A test tool", unmarshaled.Description)
	assert.Equal(t, "object", unmarshaled.InputSchema.Type)
	assert.Contains(t, unmarshaled.InputSchema.Properties, "param1")
	assert.Equal(t, "string", unmarshaled.InputSchema.Properties["param1"].Type)
	assert.Equal(t, "A test parameter", unmarshaled.InputSchema.Properties["param1"].Description)
	assert.Contains(t, unmarshaled.InputSchema.Required, "param1")
}

func TestPromptModels(t *testing.T) {
	// given
	prompt := mcp.Prompt{
		Name:        "Test Prompt",
		Description: "A test prompt",
	}

	// when we marshal and unmarshal the prompt
	data, err := json.Marshal(prompt)
	require.NoError(t, err)

	var unmarshaled mcp.Prompt
	err = json.Unmarshal(data, &unmarshaled)

	// then the prompt is preserved correctly
	require.NoError(t, err)
	assert.Equal(t, "Test Prompt", unmarshaled.Name)
	assert.Equal(t, "A test prompt", unmarshaled.Description)
}

func TestResourceModels(t *testing.T) {
	// given
	description := "A test resource"
	mimeType := "text/plain"
	resource := mcp.Resource{
		Uri:         "test://resource",
		Name:        "Test Resource",
		Description: description,
		MimeType:    mimeType,
	}

	// when we marshal and unmarshal the resource
	data, err := json.Marshal(resource)
	require.NoError(t, err)

	var unmarshaled mcp.Resource
	err = json.Unmarshal(data, &unmarshaled)

	// then the resource is preserved correctly
	require.NoError(t, err)
	assert.Equal(t, "test://resource", unmarshaled.Uri)
	assert.Equal(t, "Test Resource", unmarshaled.Name)
	assert.Equal(t, "A test resource", unmarshaled.Description)
	assert.Equal(t, "text/plain", unmarshaled.MimeType)
}

func TestCapabilitiesModels(t *testing.T) {
	// given
	serverCapabilities := mcp.ServerCapabilities{
		Tools:     &mcp.ServerToolsCapabilities{},
		Prompts:   &mcp.ServerCapabilitiesPrompts{},
		Resources: &mcp.ServerCapabilitiesResources{},
	}

	clientCapabilities := mcp.ClientCapabilities{}

	// when we marshal and unmarshal the capabilities
	serverData, err := json.Marshal(serverCapabilities)
	require.NoError(t, err)

	var unmarshaledServer mcp.ServerCapabilities
	err = json.Unmarshal(serverData, &unmarshaledServer)
	require.NoError(t, err)

	clientData, err := json.Marshal(clientCapabilities)
	require.NoError(t, err)

	var unmarshaledClient mcp.ClientCapabilities
	err = json.Unmarshal(clientData, &unmarshaledClient)

	// then the capabilities are preserved correctly
	require.NoError(t, err)
	assert.NotNil(t, unmarshaledServer.Tools)
	assert.NotNil(t, unmarshaledServer.Prompts)
	assert.NotNil(t, unmarshaledServer.Resources)
}
