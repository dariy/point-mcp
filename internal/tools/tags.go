package tools

import (
	"context"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type createTagInput struct {
	Name       string `json:"name" jsonschema:"tag display name"`
	ParentSlug string `json:"parent_slug,omitempty" jsonschema:"optional slug of the parent tag"`
}

type tagSlugInput struct {
	Slug string `json:"slug" jsonschema:"tag slug"`
}

type updateTagInput struct {
	Slug          string  `json:"slug" jsonschema:"tag slug to update"`
	Name          string  `json:"name,omitempty" jsonschema:"new display name"`
	NewSlug       string  `json:"new_slug,omitempty" jsonschema:"new URL slug"`
	ParentSlug    *string `json:"parent_slug,omitempty" jsonschema:"slug of the parent tag; use null to remove parent"`
	IsHiddenPosts *bool   `json:"is_hidden_posts,omitempty" jsonschema:"whether posts with this tag should be hidden from listings"`
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
		Name:        "point_get_tag",
		Description: "Fetch details for a single tag by slug, including its parent and children.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in tagSlugInput) (*mcp.CallToolResult, point.TagDetail, error) {
		tag, err := c.GetTag(in.Slug)
		return nil, tag, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_create_tag",
		Description: "Create a new tag.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in createTagInput) (*mcp.CallToolResult, point.TagDetail, error) {
		tag, err := c.CreateTag(point.CreateTagRequest{
			Name:       in.Name,
			ParentSlug: in.ParentSlug,
		})
		return nil, tag, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_update_tag",
		Description: "Update tag properties. Use parent_slug to organize tags into a hierarchy.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in updateTagInput) (*mcp.CallToolResult, point.TagDetail, error) {
		tag, err := c.UpdateTag(in.Slug, point.UpdateTagRequest{
			Name:          in.Name,
			Slug:          in.NewSlug,
			ParentSlug:    in.ParentSlug,
			IsHiddenPosts: in.IsHiddenPosts,
		})
		return nil, tag, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_delete_tag",
		Description: "Delete a tag. Posts using this tag will lose the tag but will not be deleted.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in tagSlugInput) (*mcp.CallToolResult, struct{}, error) {
		err := c.DeleteTag(in.Slug)
		return nil, struct{}{}, err
	})
}
