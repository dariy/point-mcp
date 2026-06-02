package tools

import (
	"context"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type updateSettingsInput struct {
	Updates map[string]string `json:"updates" jsonschema:"setting key-value pairs to apply"`
}

func registerSettingsTools(s *mcp.Server, c *point.Client) {
	AddTool(s, &mcp.Tool{
		Name:        "point_get_settings",
		Description: "Return all current blog settings as a flat key-value map.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, point.Settings, error) {
		settings, err := c.GetSettings()
		return nil, settings, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_update_settings",
		Description: "Update blog settings with the provided key-value pairs. WARNING: changes apply immediately to the live site.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in updateSettingsInput) (*mcp.CallToolResult, point.Settings, error) {
		settings, err := c.UpdateSettings(in.Updates)
		return nil, settings, err
	})
}
