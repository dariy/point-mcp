package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Register wires the three read-only MCP resources into s.
func Register(s *mcp.Server, c *point.Client) {
	s.AddResource(&mcp.Resource{
		URI:         "point://context",
		Name:        "point_context",
		Description: "Blog context: base URL, title, subtitle, author, posts-per-page, active theme metadata, and content counts.",
		MIMEType:    "application/json",
	}, contextHandler(c))

	s.AddResource(&mcp.Resource{
		URI:         "point://theme/active",
		Name:        "point_theme_active",
		Description: "Active theme CSS custom properties as a ready-to-use :root block and a key-value map.",
		MIMEType:    "application/json",
	}, themeCSSHandler(c))

	s.AddResource(&mcp.Resource{
		URI:         "point://posts/recent",
		Name:        "point_posts_recent",
		Description: "Most recent published posts (first page, default page size).",
		MIMEType:    "application/json",
	}, recentPostsHandler(c))
}

func jsonText(uri string, v any) (*mcp.ReadResourceResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(b),
		}},
	}, nil
}

// context resource

type contextPayload struct {
	BaseURL      string      `json:"base_url"`
	BlogTitle    string      `json:"blog_title"`
	Subtitle     string      `json:"subtitle"`
	AuthorName   string      `json:"author_name"`
	PostsPerPage int         `json:"posts_per_page"`
	ActiveTheme  themeInfo   `json:"active_theme"`
	Stats        statsInfo   `json:"stats"`
}

type themeInfo struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	PreviewColor string `json:"preview_color"`
	HasDarkMode  bool   `json:"has_dark_mode"`
}

type statsInfo struct {
	PublishedPosts int64 `json:"published_posts"`
	TotalPosts     int64 `json:"total_posts"`
	TotalTags      int64 `json:"total_tags"`
	TotalMedia     int64 `json:"total_media"`
}

func contextHandler(c *point.Client) mcp.ResourceHandler {
	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		type settingsCh struct {
			v   point.Settings
			err error
		}
		type themeCh struct {
			v   point.Theme
			err error
		}
		type statsCh struct {
			v   point.Stats
			err error
		}

		sCh := make(chan settingsCh, 1)
		tCh := make(chan themeCh, 1)
		stCh := make(chan statsCh, 1)

		go func() { v, err := c.GetPublicSettings(); sCh <- settingsCh{v, err} }()
		go func() { v, err := c.GetActiveTheme(); tCh <- themeCh{v, err} }()
		go func() { v, err := c.GetStats(); stCh <- statsCh{v, err} }()

		sr := <-sCh
		tr := <-tCh
		str := <-stCh

		if sr.err != nil {
			return nil, fmt.Errorf("get settings: %w", sr.err)
		}
		if tr.err != nil {
			return nil, fmt.Errorf("get active theme: %w", tr.err)
		}
		if str.err != nil {
			return nil, fmt.Errorf("get stats: %w", str.err)
		}

		postsPerPage, _ := strconv.Atoi(sr.v["posts_per_page"])

		return jsonText(req.Params.URI, contextPayload{
			BaseURL:      sr.v["url"],
			BlogTitle:    sr.v["title"],
			Subtitle:     sr.v["subtitle"],
			AuthorName:   sr.v["author_name"],
			PostsPerPage: postsPerPage,
			ActiveTheme: themeInfo{
				Name:         tr.v.Name,
				Description:  tr.v.Description,
				PreviewColor: tr.v.PreviewColor,
				HasDarkMode:  tr.v.HasDarkMode,
			},
			Stats: statsInfo{
				PublishedPosts: str.v.PublishedPosts,
				TotalPosts:     str.v.TotalPosts,
				TotalTags:      str.v.TotalTags,
				TotalMedia:     str.v.TotalMedia,
			},
		})
	}
}

// theme CSS resource

type themeCSSPayload struct {
	Variables map[string]string `json:"variables"`
	CSS       string            `json:"css"`
}

func themeCSSHandler(c *point.Client) mcp.ResourceHandler {
	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		theme, err := c.GetActiveTheme()
		if err != nil {
			return nil, fmt.Errorf("get active theme: %w", err)
		}

		colorScheme := "light"
		if theme.HasDarkMode {
			colorScheme = "dark"
		}

		vars := map[string]string{
			"color-accent": theme.PreviewColor,
			"color-scheme": colorScheme,
		}

		var sb strings.Builder
		sb.WriteString(":root {\n")
		fmt.Fprintf(&sb, "  --color-accent: %s;\n", theme.PreviewColor)
		fmt.Fprintf(&sb, "  --color-scheme: %s;\n", colorScheme)
		sb.WriteString("}\n")

		return jsonText(req.Params.URI, themeCSSPayload{
			Variables: vars,
			CSS:       sb.String(),
		})
	}
}

// recent posts resource

func recentPostsHandler(c *point.Client) mcp.ResourceHandler {
	return func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		list, err := c.ListPosts(point.PostFilter{Status: "published"})
		if err != nil {
			return nil, fmt.Errorf("list posts: %w", err)
		}
		return jsonText(req.Params.URI, list)
	}
}
