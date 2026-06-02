package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerGuidelinesTools(s *mcp.Server) {
	AddTool(s, &mcp.Tool{
		Name:        "point_get_syntax_guidelines",
		Description: "Returns detailed guidelines for authoring content in Point CMS, including extended markdown syntax, allowed HTML tags, and restricted CSS properties.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ noArgs) (*mcp.CallToolResult, syntaxGuidelinesResult, error) {
		return nil, syntaxGuidelinesResult{
			Markdown: markdownGuidelines{
				Base: "GitHub Flavored Markdown (GFM)",
				Extensions: []string{
					"Attributes: Add classes and IDs using {.class #id} syntax after elements like links or images.",
					"Fenced Divs: Wrap content in ::: {.class} ... ::: to create structural containers (render as <div class=\"class\">).",
					"Bare Media: A local media path (e.g., /2026/06/hero.jpg) on its own line automatically renders as an <img>, <video>, or <audio> tag.",
					"Auto Headings: Headers automatically get IDs for deep linking.",
					"Syntax Highlighting: Code blocks support language-specific highlighting.",
				},
			},
			HTML: htmlGuidelines{
				Policy: "bluemonday (strict whitelist)",
				AllowedTags: []string{
					"Text: br, h1-h6, p, span, em, strong, i, b, u, s, del, ins, mark",
					"Lists/Code: ul, ol, li, blockquote, code, pre, hr",
					"Layout: header, section, div, article, aside, main, nav, figure, figcaption",
					"Links: a (href, title, target, rel)",
					"Images: img (src, alt, title, width, height, loading)",
					"Video: video (src, controls, autoplay, muted, loop, playsinline, poster, preload, width, height)",
					"Audio: audio (src, controls, autoplay, loop, preload), source (src, type)",
					"SVG: svg, g, path, circle, rect, line, polyline, polygon, ellipse, text, tspans (with standard geometry and fill/stroke attrs)",
				},
				GlobalAttributes: []string{"class", "id", "role", "aria-*"},
				AllowedInlineStyles: []string{
					"color", "background-color", "background",
					"font-size", "font-weight", "font-style", "font-family", "font-variant",
					"text-align", "text-decoration", "text-transform", "text-indent",
					"line-height", "letter-spacing", "word-spacing",
					"margin", "padding (including -top, -right, -bottom, -left variants)",
					"border", "border-radius", "border-color", "border-width", "border-style",
					"width", "max-width", "min-width", "height", "max-height", "min-height",
					"display", "flex-direction", "flex-wrap", "justify-content", "align-items", "align-self", "flex", "gap", "grid-template-columns",
					"float", "clear", "overflow (including -x, -y)", "opacity", "vertical-align", "list-style", "white-space",
				},
			},
			CSS: cssGuidelines{
				Scope: "Per-post scoped CSS field",
				RestrictedPatterns: []string{
					"@import is forbidden",
					"url() with external HTTPS URLs is forbidden (use uploaded media paths)",
					"position: fixed and position: sticky are stripped",
					"z-index is stripped",
					"content property is stripped",
					"<script> tags are forbidden",
				},
			},
		}, nil
	})
}

type syntaxGuidelinesResult struct {
	Markdown markdownGuidelines `json:"markdown"`
	HTML     htmlGuidelines     `json:"html"`
	CSS      cssGuidelines      `json:"css"`
}

type markdownGuidelines struct {
	Base       string   `json:"base"`
	Extensions []string `json:"extensions"`
}

type htmlGuidelines struct {
	Policy              string   `json:"policy"`
	AllowedTags         []string `json:"allowed_tags"`
	GlobalAttributes    []string `json:"global_attributes"`
	AllowedInlineStyles []string `json:"allowed_inline_styles"`
}

type cssGuidelines struct {
	Scope              string   `json:"scope"`
	RestrictedPatterns []string `json:"restricted_patterns"`
}
