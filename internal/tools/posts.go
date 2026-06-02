package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listPostsInput struct {
	Page    int    `json:"page,omitempty" jsonschema:"page number"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"items per page"`
	Status  string `json:"status,omitempty" jsonschema:"filter by status: draft, published, scheduled"`
	Tag     string `json:"tag,omitempty" jsonschema:"filter by tag slug"`
	Search  string `json:"search,omitempty" jsonschema:"full-text search query"`
	Type    string `json:"type,omitempty" jsonschema:"post type: post, page"`
	Sort    string `json:"sort,omitempty" jsonschema:"sort order: views_desc, created_at_desc, etc."`
}

type getPostInput struct {
	ID   int64  `json:"id,omitempty" jsonschema:"post ID"`
	Slug string `json:"slug,omitempty" jsonschema:"post slug (used if id is 0)"`
}

type createPostInput struct {
	Title           string   `json:"title" jsonschema:"post title"`
	Content         string   `json:"content" jsonschema:"post body; use formatter to indicate markdown or html"`
	CSS             string   `json:"css,omitempty" jsonschema:"custom CSS injected into the post page"`
	ImmersiveMode   string   `json:"immersive_mode,omitempty" jsonschema:"immersive layout variant"`
	Formatter       string   `json:"formatter,omitempty" jsonschema:"content format: markdown or html"`
	Status          string   `json:"status,omitempty" jsonschema:"draft (default), published, or scheduled"`
	Excerpt         string   `json:"excerpt,omitempty" jsonschema:"short summary shown in listings"`
	Slug            string   `json:"slug,omitempty" jsonschema:"URL slug; auto-generated from title if omitted"`
	IsFeatured      bool     `json:"is_featured,omitempty" jsonschema:"pin post to featured slot"`
	ThumbnailPath   string   `json:"thumbnail_path,omitempty" jsonschema:"media path returned by point_upload_media"`
	MetaDescription string   `json:"meta_description,omitempty" jsonschema:"SEO meta description"`
	Tags            []string `json:"tags,omitempty" jsonschema:"list of tag slugs to attach"`
	ScheduledAt     *string  `json:"scheduled_at,omitempty" jsonschema:"RFC3339 publish time for scheduled posts"`
}

type updatePostInput struct {
	ID              int64     `json:"id" jsonschema:"post ID to update"`
	Title           *string   `json:"title,omitempty"`
	Content         *string   `json:"content,omitempty"`
	CSS             *string   `json:"css,omitempty" jsonschema:"custom CSS injected into the post page"`
	ImmersiveMode   *string   `json:"immersive_mode,omitempty"`
	Formatter       *string   `json:"formatter,omitempty" jsonschema:"markdown or html"`
	Status          *string   `json:"status,omitempty"`
	Excerpt         *string   `json:"excerpt,omitempty"`
	Slug            *string   `json:"slug,omitempty"`
	IsFeatured      *bool     `json:"is_featured,omitempty"`
	ThumbnailPath   *string   `json:"thumbnail_path,omitempty"`
	MetaDescription *string   `json:"meta_description,omitempty"`
	Tags            *[]string `json:"tags,omitempty"`
	ScheduledAt     *string   `json:"scheduled_at,omitempty"`
}

type postIDInput struct {
	ID int64 `json:"id" jsonschema:"post ID"`
}

type previewLinkResult struct {
	URL string `json:"url"`
}

type setPostTagsInput struct {
	ID   int64    `json:"id" jsonschema:"post ID"`
	Tags []string `json:"tags" jsonschema:"full list of tag slugs to set (replaces existing tags)"`
}

type replaceInPostInput struct {
	ID                int64  `json:"id" jsonschema:"post ID"`
	Field             string `json:"field" jsonschema:"field to update: content, css, or excerpt"`
	OldString         string `json:"old_string" jsonschema:"exact literal text to find"`
	NewString         string `json:"new_string" jsonschema:"replacement text"`
	AllowMultiple     bool   `json:"allow_multiple,omitempty" jsonschema:"if true, replace all occurrences; if false, fail on multiple matches"`
	ExpectedUpdatedAt string `json:"expected_updated_at,omitempty" jsonschema:"ISO timestamp for optimistic locking; fail if post has been modified since"`
}

type postWriteResult struct {
	Post        point.Post `json:"post"`
	CSSWarnings []string   `json:"css_warnings"`
}

func registerPostTools(s *mcp.Server, c *point.Client) {
	AddTool(s, &mcp.Tool{
		Name:        "point_list_posts",
		Description: "List posts with optional filters for status, tag, search query, type, and sort order.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in listPostsInput) (*mcp.CallToolResult, point.PostList, error) {
		list, err := c.ListPosts(point.PostFilter{
			Page:    in.Page,
			PerPage: in.PerPage,
			Status:  in.Status,
			Tag:     in.Tag,
			Search:  in.Search,
			Type:    in.Type,
			Sort:    in.Sort,
		})
		return nil, list, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_get_post",
		Description: "Fetch a single post by ID or slug.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in getPostInput) (*mcp.CallToolResult, point.Post, error) {
		if in.ID != 0 {
			p, err := c.GetPostByID(in.ID)
			return nil, p, err
		}
		if in.Slug != "" {
			p, err := c.GetPostBySlug(in.Slug)
			return nil, p, err
		}
		return nil, point.Post{}, fmt.Errorf("provide id or slug")
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_create_post",
		Description: "Create a new post. Default status is draft. Call point_get_syntax_guidelines first for details on markdown extensions, allowed HTML, and CSS restrictions. css_warnings in the result flag invalid CSS rules.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in createPostInput) (*mcp.CallToolResult, postWriteResult, error) {
		status := in.Status
		if status == "" {
			status = "draft"
		}
		p, warnings, err := c.CreatePost(point.CreatePostRequest{
			Title:           in.Title,
			Content:         in.Content,
			CSS:             in.CSS,
			ImmersiveMode:   in.ImmersiveMode,
			Formatter:       in.Formatter,
			Status:          status,
			Excerpt:         in.Excerpt,
			Slug:            in.Slug,
			IsFeatured:      in.IsFeatured,
			ThumbnailPath:   in.ThumbnailPath,
			MetaDescription: in.MetaDescription,
			Tags:            in.Tags,
			ScheduledAt:     in.ScheduledAt,
		})
		return nil, postWriteResult{Post: p, CSSWarnings: warnings}, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_update_post",
		Description: "Update an existing post by ID. Call point_get_syntax_guidelines first for details on markdown extensions, allowed HTML, and CSS restrictions. Only provided fields are changed; omitted fields retain their current values.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in updatePostInput) (*mcp.CallToolResult, postWriteResult, error) {
		// 1. Fetch current post to perform non-destructive update (backend only supports PUT).
		current, err := c.GetPostByID(in.ID)
		if err != nil {
			return nil, postWriteResult{}, fmt.Errorf("fetching post for update: %w", err)
		}

		// 2. Prepare update request starting with current values.
		req := point.UpdatePostRequest{
			Title:         current.Title,
			Content:       current.Content,
			CSS:           current.CSS,
			ImmersiveMode: current.ImmersiveMode,
			Formatter:     current.Formatter,
			Status:        current.Status,
			Slug:          current.Slug,
			IsFeatured:    &current.IsFeatured,
		}

		if current.MetaDescription != nil {
			req.MetaDescription = *current.MetaDescription
		}
		if current.Excerpt != nil {
			req.Excerpt = *current.Excerpt
		}
		if current.ThumbnailPath != nil {
			req.ThumbnailPath = *current.ThumbnailPath
		}
		if current.ScheduledAt != nil {
			s := current.ScheduledAt.Format(time.RFC3339)
			req.ScheduledAt = &s
		}

		// Convert current tags back to slug list.
		tags := make([]string, len(current.Tags))
		for i, t := range current.Tags {
			tags[i] = t.Slug
		}
		req.Tags = tags

		// 3. Override with provided fields.
		if in.Title != nil {
			req.Title = *in.Title
		}
		if in.Content != nil {
			req.Content = *in.Content
		}
		if in.CSS != nil {
			req.CSS = *in.CSS
		}
		if in.ImmersiveMode != nil {
			req.ImmersiveMode = *in.ImmersiveMode
		}
		if in.Formatter != nil {
			req.Formatter = *in.Formatter
		}
		if in.Status != nil {
			req.Status = *in.Status
		}
		if in.Excerpt != nil {
			req.Excerpt = *in.Excerpt
		}
		if in.Slug != nil {
			req.Slug = *in.Slug
		}
		if in.IsFeatured != nil {
			req.IsFeatured = in.IsFeatured
		}
		if in.ThumbnailPath != nil {
			req.ThumbnailPath = *in.ThumbnailPath
		}
		if in.MetaDescription != nil {
			req.MetaDescription = *in.MetaDescription
		}
		if in.Tags != nil {
			req.Tags = *in.Tags
		}
		if in.ScheduledAt != nil {
			req.ScheduledAt = in.ScheduledAt
		}

		p, warnings, err := c.UpdatePost(in.ID, req)
		return nil, postWriteResult{Post: p, CSSWarnings: warnings}, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_publish_post",
		Description: "Publish a post immediately, making it visible on the live site.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in postIDInput) (*mcp.CallToolResult, point.Post, error) {
		p, err := c.Publish(in.ID)
		return nil, p, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_withdraw_post",
		Description: "Withdraw (unpublish) a post, reverting it to draft status.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in postIDInput) (*mcp.CallToolResult, point.Post, error) {
		p, err := c.Withdraw(in.ID)
		return nil, p, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_delete_post",
		Description: "Permanently delete a post. WARNING: this action cannot be undone.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in postIDInput) (*mcp.CallToolResult, struct{}, error) {
		err := c.DeletePost(in.ID)
		return nil, struct{}{}, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_generate_preview_link",
		Description: "Generate a temporary preview URL for a draft post.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, func(_ context.Context, _ *mcp.CallToolRequest, in postIDInput) (*mcp.CallToolResult, previewLinkResult, error) {
		url, err := c.GeneratePreviewLink(in.ID)
		return nil, previewLinkResult{URL: url}, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_set_post_tags",
		Description: "Replace all tags on a post. The tags list is the new complete set; tags not listed are removed.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in setPostTagsInput) (*mcp.CallToolResult, point.Post, error) {
		p, err := c.UpdateTags(in.ID, in.Tags)
		return nil, p, err
	})

	AddTool(s, &mcp.Tool{
		Name:        "point_replace_in_post",
		Description: "Perform a targeted string replacement within a specific post field (content, css, or excerpt). This is more token-efficient than updating the entire content. Returns the updated post.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, in replaceInPostInput) (*mcp.CallToolResult, postWriteResult, error) {
		// 1. Fetch current post.
		current, err := c.GetPostByID(in.ID)
		if err != nil {
			return nil, postWriteResult{}, fmt.Errorf("fetching post for replacement: %w", err)
		}

		// 2. Validate optimistic locking.
		if in.ExpectedUpdatedAt != "" {
			expected, err := time.Parse(time.RFC3339, in.ExpectedUpdatedAt)
			if err != nil {
				return nil, postWriteResult{}, fmt.Errorf("invalid expected_updated_at format: %w", err)
			}
			// Compare timestamps
			if !current.UpdatedAt.Equal(expected) {
				return nil, postWriteResult{}, fmt.Errorf("optimistic locking failure: post was modified at %s, expected %s",
					current.UpdatedAt.Format(time.RFC3339), expected.Format(time.RFC3339))
			}
		}

		// 3. Get target field value.
		var currentVal string
		switch strings.ToLower(in.Field) {
		case "content":
			currentVal = current.Content
		case "css":
			currentVal = current.CSS
		case "excerpt":
			if current.Excerpt != nil {
				currentVal = *current.Excerpt
			}
		default:
			return nil, postWriteResult{}, fmt.Errorf("invalid field: %s (must be content, css, or excerpt)", in.Field)
		}

		// 4. Perform replacement logic.
		count := strings.Count(currentVal, in.OldString)
		if count == 0 {
			// Idempotency check: if new_string is already there, consider it a success.
			if strings.Contains(currentVal, in.NewString) {
				return nil, postWriteResult{Post: current}, nil
			}
			return nil, postWriteResult{}, fmt.Errorf("old_string not found in field %s", in.Field)
		}

		if !in.AllowMultiple && count > 1 {
			return nil, postWriteResult{}, fmt.Errorf("found %d occurrences of old_string, but allow_multiple is false", count)
		}

		newVal := strings.ReplaceAll(currentVal, in.OldString, in.NewString)
		if !in.AllowMultiple && count == 1 {
			newVal = strings.Replace(currentVal, in.OldString, in.NewString, 1)
		}

		// 5. Prepare update request (standard non-destructive update logic).
		req := point.UpdatePostRequest{
			Title:         current.Title,
			Content:       current.Content,
			CSS:           current.CSS,
			ImmersiveMode: current.ImmersiveMode,
			Formatter:     current.Formatter,
			Status:        current.Status,
			Slug:          current.Slug,
			IsFeatured:    &current.IsFeatured,
		}

		if current.MetaDescription != nil {
			req.MetaDescription = *current.MetaDescription
		}
		if current.Excerpt != nil {
			req.Excerpt = *current.Excerpt
		}
		if current.ThumbnailPath != nil {
			req.ThumbnailPath = *current.ThumbnailPath
		}
		if current.ScheduledAt != nil {
			s := current.ScheduledAt.Format(time.RFC3339)
			req.ScheduledAt = &s
		}

		tags := make([]string, len(current.Tags))
		for i, t := range current.Tags {
			tags[i] = t.Slug
		}
		req.Tags = tags

		// Override the target field.
		switch strings.ToLower(in.Field) {
		case "content":
			req.Content = newVal
		case "css":
			req.CSS = newVal
		case "excerpt":
			req.Excerpt = newVal
		}

		p, warnings, err := c.UpdatePost(in.ID, req)
		return nil, postWriteResult{Post: p, CSSWarnings: warnings}, err
	})
}
