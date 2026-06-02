package prompts

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCreateLandingPageHandler(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name:      "create_landing_page",
			Arguments: map[string]string{"topic": "PointBlog CMS", "media_paths": "/tmp/hero.jpg,/tmp/feature.png"},
		},
	}
	result, err := createLandingPageHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || len(result.Messages) == 0 {
		t.Fatal("expected messages")
	}
	body := result.Messages[0].Content.(*mcp.TextContent).Text
	checks := []string{
		"PointBlog CMS",
		"/tmp/hero.jpg",
		"/tmp/feature.png",
		"point_get_context",
		"point_get_theme_css",
		"point_get_syntax_guidelines",
		"point_upload_media",
		"point_create_post",
		"point_generate_preview_link",
		"::: {.classname}",
		"immersive_mode",
		"--color-accent",
		"bluemonday",
	}
	for _, want := range checks {
		if !strings.Contains(body, want) {
			t.Errorf("prompt body missing: %q", want)
		}
	}
	t.Logf("Prompt length: %d chars", len(body))
}

func TestCreateLandingPageHandlerNoMedia(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name:      "create_landing_page",
			Arguments: map[string]string{"topic": "Widget Co"},
		},
	}
	result, err := createLandingPageHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := result.Messages[0].Content.(*mcp.TextContent).Text
	if !strings.Contains(body, "Widget Co") {
		t.Error("topic not in body")
	}
}

func TestCreateLandingPageHandlerEmptyTopic(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Name:      "create_landing_page",
			Arguments: map[string]string{"topic": ""},
		},
	}
	_, err := createLandingPageHandler(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty topic")
	}
}
