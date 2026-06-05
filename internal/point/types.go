package point

import "time"

// Tag is a minimal tag object as returned in post responses.
type Tag struct {
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	IsHiddenPosts *bool  `json:"is_hidden_posts,omitempty"`
}

// Media is an attachment linked to a post.
type Media struct {
	Path     string                  `json:"path"`
	AltText  *string                 `json:"alt_text"`
	Metadata *map[string]interface{} `json:"metadata"`
}

// Post is the API response shape for a single post or list item.
// Detail-only fields (ContentHTML, Media) are nil/empty for list responses.
// List-only fields (MediaURL) are nil for detail responses.
type Post struct {
	ID              int64      `json:"id"`
	Title           string     `json:"title"`
	Slug            string     `json:"slug"`
	Type            string     `json:"type"`
	Content         string     `json:"content"`
	ContentHTML     string     `json:"content_html,omitempty"`
	CSS             string     `json:"css"`
	ImmersiveMode   string     `json:"immersive_mode"`
	Excerpt         *string    `json:"excerpt"`
	Formatter       string     `json:"formatter"`
	Status          string     `json:"status"`
	IsFeatured      bool       `json:"is_featured"`
	ViewCount       int64      `json:"view_count"`
	PublishedAt     *time.Time `json:"published_at"`
	ScheduledAt     *time.Time `json:"scheduled_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ThumbnailPath   *string    `json:"thumbnail_path,omitempty"`
	MediaURL        *string    `json:"media_url,omitempty"`
	MetaDescription *string    `json:"meta_description"`
	Tags            []Tag      `json:"tags"`
	Media           []Media    `json:"media,omitempty"`

	// Admin-only fields injected by the server.
	IsHidden      *bool `json:"is_hidden,omitempty"`
	IsHiddenByTag *bool `json:"is_hidden_by_tag,omitempty"`
}

// PostList is the paginated list response from GET /posts.
type PostList struct {
	Posts   []Post `json:"posts"`
	Total   int64  `json:"total"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Pages   int    `json:"pages"`
}

// CreatePostRequest is the body for POST /posts, mirroring the server's CreatePostRequest.
type CreatePostRequest struct {
	Title           string   `json:"title"`
	Content         string   `json:"content"`
	CSS             string   `json:"css,omitempty"`
	ImmersiveMode   string   `json:"immersive_mode,omitempty"`
	Excerpt         string   `json:"excerpt,omitempty"`
	Slug            string   `json:"slug,omitempty"`
	Formatter       string   `json:"formatter,omitempty"`
	Status          string   `json:"status,omitempty"`
	IsFeatured      bool     `json:"is_featured,omitempty"`
	ThumbnailPath   string   `json:"thumbnail_path,omitempty"`
	MetaDescription string   `json:"meta_description,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	ScheduledAt     *string  `json:"scheduled_at,omitempty"`
}

// UpdatePostRequest is the body for PUT /posts/:id.
type UpdatePostRequest struct {
	Title           string   `json:"title,omitempty"`
	Content         string   `json:"content,omitempty"`
	CSS             string   `json:"css,omitempty"`
	ImmersiveMode   string   `json:"immersive_mode,omitempty"`
	Excerpt         string   `json:"excerpt,omitempty"`
	Slug            string   `json:"slug,omitempty"`
	Formatter       string   `json:"formatter,omitempty"`
	Status          string   `json:"status,omitempty"`
	IsFeatured      *bool    `json:"is_featured,omitempty"`
	ThumbnailPath   string   `json:"thumbnail_path,omitempty"`
	MetaDescription string   `json:"meta_description,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	ScheduledAt     *string  `json:"scheduled_at,omitempty"`
}

// Settings is a flat map of all blog settings, keyed by setting name.
type Settings map[string]string

// Theme describes an available UI theme.
type Theme struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	PreviewColor string `json:"preview_color"`
	HasDarkMode  bool   `json:"has_dark_mode"`
}

// Stats is the response from GET /system/stats.
type Stats struct {
	PublishedPosts   int64   `json:"published_posts"`
	TotalPosts       int64   `json:"total_posts"`
	TotalTags        int64   `json:"total_tags"`
	TotalMedia       int64   `json:"total_media"`
	StorageUsedMB    float64 `json:"storage_used_mb"`
	UptimeSeconds    int64   `json:"uptime_seconds"`
	ImportConfigured bool    `json:"import_configured"`
}

// MediaItem is the full media object returned by the media API endpoints.
type MediaItem struct {
	ID          int64                   `json:"id"`
	Filename    string                  `json:"filename"`
	Path        string                  `json:"path"`
	URL         string                  `json:"url"`
	Size        int64                   `json:"size"`
	ContentType string                  `json:"content_type"`
	AltText     *string                 `json:"alt_text"`
	Metadata    *map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time               `json:"created_at"`
}

// MediaList is the paginated list response from GET /media.
type MediaList struct {
	Media   []MediaItem `json:"media"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
	Pages   int         `json:"pages"`
}

// MediaAnalysis is the response from POST /media/:id/analyze.
type MediaAnalysis struct {
	Title   string   `json:"title"`
	Tags    []string `json:"tags"`
	Excerpt string   `json:"excerpt"`
}

// TagDetail is the full tag object returned by the tags API endpoints.
type TagDetail struct {
	Name          string       `json:"name"`
	Slug          string       `json:"slug"`
	PostCount     int64        `json:"post_count"`
	IsHiddenPosts *bool        `json:"is_hidden_posts,omitempty"`
	Parent        *TagSummary  `json:"parent,omitempty"`
	Children      []TagSummary `json:"children,omitempty"`
}

// TagSummary is a simplified tag object used in parent/children fields to avoid recursion.
type TagSummary struct {
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	PostCount     int64  `json:"post_count"`
	IsHiddenPosts *bool  `json:"is_hidden_posts,omitempty"`
}

// CreateTagRequest is the body for POST /tags.
type CreateTagRequest struct {
	Name       string `json:"name"`
	ParentSlug string `json:"parent_slug,omitempty"`
}

// UpdateTagRequest is the body for PUT /tags/:slug.
type UpdateTagRequest struct {
	Name          string  `json:"name,omitempty"`
	Slug          string  `json:"slug,omitempty"`
	ParentSlug    *string `json:"parent_slug,omitempty"`
	IsHiddenPosts *bool   `json:"is_hidden_posts,omitempty"`
}

// StorageStats is the response from GET /media/storage-stats.
type StorageStats struct {
	TotalBytes int64 `json:"total_bytes"`
	TotalFiles int64 `json:"total_files"`
	ImageCount int64 `json:"image_count"`
	VideoCount int64 `json:"video_count"`
	AudioCount int64 `json:"audio_count"`
	OtherCount int64 `json:"other_count"`
}

// PostAnalytics is the response from GET /api/analytics/posts.
type PostAnalytics struct {
	TotalViews          int64   `json:"total_views"`
	AverageViewsPerPost float64 `json:"average_views_per_post"`
	MostViewedPostID    int64   `json:"most_viewed_post_id"`
}

// ErrorResponse is the standard error body returned by the API.
type ErrorResponse struct {
	Detail string `json:"detail"`
}
