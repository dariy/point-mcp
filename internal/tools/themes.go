package tools

import (
	"context"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type setActiveThemeInput struct {
	Name string `json:"name" jsonschema:"name of the theme to activate"`
}

type listThemesResult struct {
	Themes []point.Theme `json:"themes"`
}

func registerThemeTools(s *mcp.Server, c *point.Client) {
	AddTool(s, &mcp.Tool{
		Name:        "point_list_themes",
		Description: "List all themes available for the blog.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listThemesResult, error) {
		themes, err := c.ListThemes()
		return nil, listThemesResult{Themes: themes}, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_set_active_theme",
		Description: "Activate a theme by name. WARNING: this changes the live site's appearance immediately.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in setActiveThemeInput) (*mcp.CallToolResult, point.Theme, error) {
		theme, err := c.SetActiveTheme(in.Name)
		return nil, theme, err
	})
}
