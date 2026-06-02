package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerContextTools(srv *mcp.Server, c *point.Client) {
	AddTool(srv, &mcp.Tool{
		Name:        "point_get_context",
		Description: "Returns blog context needed to write content: base URL, title, subtitle, author name, posts per page, active theme metadata, and content count stats.",
	}, getContextHandler(c))

	AddTool(srv, &mcp.Tool{
		Name:        "point_get_theme_css",
		Description: "Returns the active theme's CSS custom properties so generated post CSS harmonizes with the blog theme.",
	}, getThemeCSSHandler(c))
}

// noArgs is the input type for tools that accept no parameters.
type noArgs struct{}

// contextResult is the structured output of point_get_context.
type contextResult struct {
	BaseURL      string    `json:"base_url"       jsonschema:"blog base URL"`
	BlogTitle    string    `json:"blog_title"     jsonschema:"blog title"`
	Subtitle     string    `json:"subtitle"       jsonschema:"blog subtitle"`
	AuthorName   string    `json:"author_name"    jsonschema:"primary author name"`
	PostsPerPage int       `json:"posts_per_page" jsonschema:"configured posts-per-page"`
	ActiveTheme  themeInfo `json:"active_theme"   jsonschema:"metadata for the currently active theme"`
	Stats        statsInfo `json:"stats"          jsonschema:"published content counts"`
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

// themeCSSResult is the structured output of point_get_theme_css.
type themeCSSResult struct {
	Variables map[string]string `json:"variables" jsonschema:"CSS variable names (without leading --) mapped to their values"`
	CSS       string            `json:"css"       jsonschema:"ready-to-use CSS :root block containing the variables"`
}

func getContextHandler(c *point.Client) mcp.ToolHandlerFor[*noArgs, *contextResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ *noArgs) (*mcp.CallToolResult, *contextResult, error) {
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
			return nil, nil, fmt.Errorf("get settings: %w", sr.err)
		}
		if tr.err != nil {
			return nil, nil, fmt.Errorf("get active theme: %w", tr.err)
		}
		if str.err != nil {
			return nil, nil, fmt.Errorf("get stats: %w", str.err)
		}

		postsPerPage, _ := strconv.Atoi(sr.v["posts_per_page"])

		return nil, &contextResult{
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
		}, nil
	}
}

func getThemeCSSHandler(c *point.Client) mcp.ToolHandlerFor[*noArgs, *themeCSSResult] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ *noArgs) (*mcp.CallToolResult, *themeCSSResult, error) {
		theme, err := c.GetActiveTheme()
		if err != nil {
			return nil, nil, fmt.Errorf("get active theme: %w", err)
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
		// deterministic order: accent before scheme
		fmt.Fprintf(&sb, "  --color-accent: %s;\n", theme.PreviewColor)
		fmt.Fprintf(&sb, "  --color-scheme: %s;\n", colorScheme)
		sb.WriteString("}\n")

		return nil, &themeCSSResult{
			Variables: vars,
			CSS:       sb.String(),
		}, nil
	}
}
