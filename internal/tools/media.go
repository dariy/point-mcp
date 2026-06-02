package tools

import (
	"context"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type uploadMediaInput struct {
	FilePath string `json:"file_path" jsonschema:"absolute path to the local file to upload"`
}

type listMediaInput struct {
	Page    int `json:"page,omitempty" jsonschema:"page number"`
	PerPage int `json:"per_page,omitempty" jsonschema:"items per page"`
}

type analyzeImageInput struct {
	ID int64 `json:"id" jsonschema:"media item ID to analyze"`
}

func registerMediaTools(s *mcp.Server, c *point.Client) {
	AddTool(s, &mcp.Tool{
		Name:        "point_upload_media",
		Description: "Upload a local file to the media library. Returns the media item including its path (e.g. /2024/01/image.jpg) for use as thumbnail_path or embedded in post content.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in uploadMediaInput) (*mcp.CallToolResult, point.MediaItem, error) {
		item, err := c.UploadFile(in.FilePath)
		return nil, item, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_list_media",
		Description: "List media items in the library with optional pagination.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in listMediaInput) (*mcp.CallToolResult, point.MediaList, error) {
		list, err := c.ListMedia(point.MediaFilter{
			Page:    in.Page,
			PerPage: in.PerPage,
		})
		return nil, list, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_analyze_media",
		Description: "Analyze a media image using AI to generate a suggested title, tags, and excerpt.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in analyzeImageInput) (*mcp.CallToolResult, point.MediaAnalysis, error) {
		analysis, err := c.AnalyzeImageByID(in.ID)
		return nil, analysis, err
	})
}
