package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/mcpshim/mcpshim/internal/config"
	"github.com/mcpshim/mcpshim/internal/protocol"
)

func TestParseSchema(t *testing.T) {
	schema := map[string]interface{}{
		"required": []interface{}{"query"},
		"properties": map[string]interface{}{
			"query":  map[string]interface{}{"type": "string"},
			"limit":  map[string]interface{}{"type": "integer"},
			"filter": map[string]interface{}{"type": "string"},
		},
	}

	required, props := parseSchema(schema)

	if len(required) != 1 || required[0] != "query" {
		t.Errorf("expected required=[query], got %v", required)
	}
	if len(props) != 3 {
		t.Errorf("expected 3 properties, got %d", len(props))
	}
	// properties should be sorted
	if props[0] != "filter" || props[1] != "limit" || props[2] != "query" {
		t.Errorf("expected sorted properties [filter limit query], got %v", props)
	}
}

func TestParseSchemaDetail(t *testing.T) {
	schema := map[string]interface{}{
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Max results",
			},
		},
	}
	required := []string{"query"}

	details := parseSchemaDetail(schema, required)

	if len(details) != 2 {
		t.Fatalf("expected 2 property details, got %d", len(details))
	}

	// details are sorted alphabetically: limit then query
	limitDetail := details[0]
	if limitDetail.Name != "limit" {
		t.Errorf("expected first detail name=limit, got %s", limitDetail.Name)
	}
	if limitDetail.Type != "integer" {
		t.Errorf("expected limit type=integer, got %s", limitDetail.Type)
	}
	if limitDetail.Required {
		t.Error("expected limit to not be required")
	}
	if limitDetail.Description != "Max results" {
		t.Errorf("expected limit description='Max results', got %s", limitDetail.Description)
	}

	queryDetail := details[1]
	if queryDetail.Name != "query" {
		t.Errorf("expected second detail name=query, got %s", queryDetail.Name)
	}
	if queryDetail.Type != "string" {
		t.Errorf("expected query type=string, got %s", queryDetail.Type)
	}
	if !queryDetail.Required {
		t.Error("expected query to be required")
	}
	if queryDetail.Description != "Search query" {
		t.Errorf("expected query description='Search query', got %s", queryDetail.Description)
	}
}

func TestParseSchemaDetailEmpty(t *testing.T) {
	details := parseSchemaDetail(nil, nil)
	if len(details) != 0 {
		t.Errorf("expected empty result for nil schema, got %v", details)
	}
}

func TestParseSchemaEmpty(t *testing.T) {
	required, props := parseSchema(map[string]interface{}{})
	if len(required) != 0 {
		t.Errorf("expected no required fields, got %v", required)
	}
	if len(props) != 0 {
		t.Errorf("expected no properties, got %v", props)
	}
}

func TestLoginRejectsStdioServer(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.MCPServer{
			{Name: "local", Transport: "stdio", Command: []string{"echo"}},
		},
	}
	reg := NewRegistry(cfg, nil)
	err := reg.Login(context.Background(), "local", false)
	if err == nil {
		t.Fatal("expected error for stdio login, got nil")
	}
	if !strings.Contains(err.Error(), "stdio") {
		t.Errorf("expected error to mention stdio, got: %s", err.Error())
	}
}

func TestLoginRejectsUnknownServer(t *testing.T) {
	cfg := &config.Config{}
	reg := NewRegistry(cfg, nil)
	err := reg.Login(context.Background(), "nonexistent", false)
	if err == nil {
		t.Fatal("expected error for unknown server, got nil")
	}
	if !strings.Contains(err.Error(), "unknown server") {
		t.Errorf("expected 'unknown server' error, got: %s", err.Error())
	}
}

func TestNewClientRejectsEmptyCommand(t *testing.T) {
	s := config.MCPServer{Name: "empty", Transport: "stdio", Command: []string{}}
	_, _, err := newClient(s)
	if err == nil {
		t.Fatal("expected error for empty command, got nil")
	}
	if !strings.Contains(err.Error(), "no command configured") {
		t.Errorf("expected 'no command configured' error, got: %s", err.Error())
	}
}

func TestNewClientRejectsNilCommand(t *testing.T) {
	s := config.MCPServer{Name: "nilcmd", Transport: "stdio"}
	_, _, err := newClient(s)
	if err == nil {
		t.Fatal("expected error for nil command, got nil")
	}
	if !strings.Contains(err.Error(), "no command configured") {
		t.Errorf("expected 'no command configured' error, got: %s", err.Error())
	}
}

func TestServersIncludesEnv(t *testing.T) {
	cfg := &config.Config{
		Servers: []config.MCPServer{
			{
				Name:      "local",
				Transport: "stdio",
				Command:   []string{"echo"},
				Env:       []string{"FOO=bar"},
			},
		},
	}
	reg := NewRegistry(cfg, nil)
	servers := reg.Servers()
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	if len(servers[0].Env) != 1 || servers[0].Env[0] != "FOO=bar" {
		t.Errorf("expected Env=[FOO=bar], got %v", servers[0].Env)
	}
	if len(servers[0].Command) != 1 || servers[0].Command[0] != "echo" {
		t.Errorf("expected Command=[echo], got %v", servers[0].Command)
	}
}

func TestToolDetailProtocol(t *testing.T) {
	d := &protocol.ToolDetail{
		Server:      "myserver",
		Name:        "search",
		Description: "Search items",
		Properties: []protocol.PropertyDetail{
			{Name: "query", Type: "string", Description: "Search query", Required: true},
			{Name: "limit", Type: "integer", Description: "Max results", Required: false},
		},
	}

	if d.Server != "myserver" {
		t.Errorf("unexpected server: %s", d.Server)
	}
	if len(d.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(d.Properties))
	}
	if !d.Properties[0].Required {
		t.Error("expected first property to be required")
	}
}
