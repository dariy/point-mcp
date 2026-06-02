package tools

import (
	"context"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type topPostsInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"number of top posts to return (default 10)"`
}

func registerAnalyticsTools(s *mcp.Server, c *point.Client) {
	AddTool(s, &mcp.Tool{
		Name:        "point_analytics_top_posts",
		Description: "Retrieve top-performing published posts sorted by view count.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in topPostsInput) (*mcp.CallToolResult, point.PostList, error) {
		limit := in.Limit
		if limit <= 0 {
			limit = 10
		}
		list, err := c.ListPosts(point.PostFilter{
			PerPage: limit,
			Status:  "published",
			Sort:    "views",
		})
		return nil, list, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_analytics_summary",
		Description: "Retrieve aggregated post statistics including total and average view counts.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, point.PostAnalytics, error) {
		stats, err := c.GetPostAnalytics()
		return nil, stats, err
	})
}
