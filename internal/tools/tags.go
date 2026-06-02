package tools

import (
	"context"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type createTagInput struct {
	Name string `json:"name" jsonschema:"tag display name"`
}

type listTagsResult struct {
	Tags []point.TagDetail `json:"tags"`
}

func registerTagTools(s *mcp.Server, c *point.Client) {
	AddTool(s, &mcp.Tool{
		Name:        "point_list_tags",
		Description: "List all tags with their slugs and post counts.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, listTagsResult, error) {
		tags, err := c.ListTags()
		return nil, listTagsResult{Tags: tags}, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_create_tag",
		Description: "Create a new tag.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in createTagInput) (*mcp.CallToolResult, point.TagDetail, error) {
		tag, err := c.CreateTag(in.Name)
		return nil, tag, err
	})
}
