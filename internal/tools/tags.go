package tools

import (
	"context"
	"fmt"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type createTagInput struct {
	Name        string              `json:"name" jsonschema:"tag display name"`
	Slug        string              `json:"slug,omitempty" jsonschema:"optional custom URL slug"`
	Description string              `json:"description,omitempty" jsonschema:"tag description/docs"`
	CustomURL   string              `json:"custom_url,omitempty" jsonschema:"external link"`
	SortOrder   *int32              `json:"sort_order,omitempty" jsonschema:"manual sorting order"`
	ParentIDs   []int64             `json:"parent_ids,omitempty" jsonschema:"list of parent tag IDs (many-to-many)"`
	ChildIDs    []int64             `json:"child_ids,omitempty" jsonschema:"list of child tag IDs (many-to-many)"`
	Locations   []point.TagLocation `json:"locations,omitempty" jsonschema:"map coordinates"`
}

type getTagInput struct {
	ID   int64  `json:"id,omitempty" jsonschema:"tag ID"`
	Slug string `json:"slug,omitempty" jsonschema:"tag slug (used if id is 0)"`
}

type updateTagInput struct {
	ID          int64               `json:"id" jsonschema:"tag ID to update"`
	Name        string              `json:"name,omitempty" jsonschema:"new display name"`
	Slug        string              `json:"slug,omitempty" jsonschema:"new URL slug"`
	Description string              `json:"description,omitempty" jsonschema:"tag description/docs"`
	CustomURL   string              `json:"custom_url,omitempty" jsonschema:"external link"`
	SortOrder   *int32              `json:"sort_order,omitempty" jsonschema:"manual sorting order"`
	ParentIDs   []int64             `json:"parent_ids,omitempty" jsonschema:"complete list of parent IDs (replaces existing)"`
	ChildIDs    []int64             `json:"child_ids,omitempty" jsonschema:"complete list of child IDs (replaces existing)"`
	Locations   []point.TagLocation `json:"locations,omitempty" jsonschema:"complete list of coordinates (replaces existing)"`
}

type tagIDInput struct {
	ID int64 `json:"id" jsonschema:"tag ID"`
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
		Description: "Fetch details for a single tag by ID or slug, including its parents, children, features, and map coordinates.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in getTagInput) (*mcp.CallToolResult, point.TagDetail, error) {
		if in.ID != 0 {
			tag, err := c.GetTag(in.ID)
			return nil, tag, err
		}
		if in.Slug != "" {
			tag, err := c.GetTagBySlug(in.Slug)
			return nil, tag, err
		}
		return nil, point.TagDetail{}, fmt.Errorf("provide id or slug")
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_create_tag",
		Description: "Create a new tag with optional description, hierarchy, and map coordinates.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in createTagInput) (*mcp.CallToolResult, point.TagDetail, error) {
		tag, err := c.CreateTag(point.CreateTagRequest{
			Name:        in.Name,
			Slug:        in.Slug,
			Description: in.Description,
			CustomURL:   in.CustomURL,
			SortOrder:   in.SortOrder,
			ParentIDs:   in.ParentIDs,
			ChildIDs:    in.ChildIDs,
			Locations:   in.Locations,
		})
		return nil, tag, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_update_tag",
		Description: "Update tag properties. Use parent_ids and child_ids to manage many-to-many hierarchy. Use locations to set map coordinates.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in updateTagInput) (*mcp.CallToolResult, point.TagDetail, error) {
		tag, err := c.UpdateTag(in.ID, point.UpdateTagRequest{
			Name:        in.Name,
			Slug:        in.Slug,
			Description: in.Description,
			CustomURL:   in.CustomURL,
			SortOrder:   in.SortOrder,
			ParentIDs:   in.ParentIDs,
			ChildIDs:    in.ChildIDs,
			Locations:   in.Locations,
		})
		return nil, tag, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_delete_tag",
		Description: "Delete a tag. Posts using this tag will lose the tag but will not be deleted.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in tagIDInput) (*mcp.CallToolResult, struct{}, error) {
		err := c.DeleteTag(in.ID)
		return nil, struct{}{}, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_geocode_tag",
		Description: "Automatically look up and store map coordinates for a tag based on its name (uses OpenStreetMap Nominatim).",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in tagIDInput) (*mcp.CallToolResult, point.TagLocation, error) {
		loc, err := c.GeocodeTag(in.ID)
		return nil, loc, err
	})
}
