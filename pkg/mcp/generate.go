package mcp

// This is for APIs, not schemas
// Xgo:generate go run -modfile=../../../../tools/go.mod github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml schema.json

// https://github.com/omissis/go-jsonschema/blob/main/main.go
//go:generate go-jsonschema -p shared schema.json -o shared/mcp_models.go
