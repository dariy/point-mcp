package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// toolRegistry stores handlers for direct CLI invocation.
var toolRegistry = make(map[string]func(context.Context, json.RawMessage) (any, error))

// Register wires all MCP tools into s using the provided Point API client.
func Register(s *mcp.Server, c *point.Client) {
	registerContextTools(s, c)
	registerSettingsTools(s, c)
	registerThemeTools(s, c)
	registerPostTools(s, c)
	registerMediaTools(s, c)
	registerTagTools(s, c)
	registerAnalyticsTools(s, c)
	registerGuidelinesTools(s)
}

func AddTool[T any, R any](s *mcp.Server, tool *mcp.Tool, handler mcp.ToolHandlerFor[T, R]) {
	mcp.AddTool(s, tool, func(ctx context.Context, req *mcp.CallToolRequest, in T) (*mcp.CallToolResult, R, error) {
		inJSON, _ := json.Marshal(in)
		displayArgs := string(inJSON)
		if len(displayArgs) > 1000 {
			displayArgs = displayArgs[:1000] + "... (truncated)"
		}
		log.Printf("[MCP Tool Call] %s | Args: %s", tool.Name, displayArgs)

		res, out, err := handler(ctx, req, in)

		if err != nil {
			log.Printf("[MCP Tool Error] %s | Error: %v", tool.Name, err)
		} else {
			outJSON, _ := json.Marshal(out)
			displayResult := string(outJSON)
			if len(displayResult) > 1000 {
				displayResult = displayResult[:1000] + "... (truncated)"
			}
			log.Printf("[MCP Tool Success] %s | Result: %s", tool.Name, displayResult)
		}

		return res, out, err
	})

	// Register for direct CLI call
	toolRegistry[tool.Name] = func(ctx context.Context, args json.RawMessage) (any, error) {
		var in T
		if len(args) > 0 && string(args) != "null" {
			if err := json.Unmarshal(args, &in); err != nil {
				return nil, err
			}
		}
		_, out, err := handler(ctx, nil, in)
		return out, err
	}
}

// Dispatch calls a tool by name with the given JSON arguments.
func Dispatch(ctx context.Context, name string, args json.RawMessage) (any, error) {
	h, ok := toolRegistry[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return h(ctx, args)
}
