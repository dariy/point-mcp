package point

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// PostFilter holds optional query parameters for ListPosts.
type PostFilter struct {
	Page    int
	PerPage int
	Status  string
	Tag     string
	Search  string
	Type    string
	Sort    string
}

func (f PostFilter) query() string {
	v := url.Values{}
	if f.Page > 0 {
		v.Set("page", strconv.Itoa(f.Page))
	}
	if f.PerPage > 0 {
		v.Set("per_page", strconv.Itoa(f.PerPage))
	}
	if f.Status != "" {
		v.Set("status", f.Status)
	}
	if f.Tag != "" {
		v.Set("tag", f.Tag)
	}
	if f.Search != "" {
		v.Set("search", f.Search)
	}
	if f.Type != "" {
		v.Set("type", f.Type)
	}
	if f.Sort != "" {
		v.Set("sort", f.Sort)
	}
	if qs := v.Encode(); qs != "" {
		return "?" + qs
	}
	return ""
}

// postWriteResponse is the API envelope for create/update that carries CSS validation warnings.
type postWriteResponse struct {
	Post
	WrappedPost Post     `json:"post"`
	CSSWarnings []string `json:"css_warnings"`
}

func (r postWriteResponse) post() Post {
	if r.WrappedPost.ID != 0 {
		return r.WrappedPost
	}
	return r.Post
}

type previewLinkResponse struct {
	URL string `json:"url"`
}

func (c *Client) ListPosts(filter PostFilter) (PostList, error) {
	return get[PostList](c, "/api/posts"+filter.query())
}

func (c *Client) GetPostByID(id int64) (Post, error) {
	return get[Post](c, fmt.Sprintf("/api/posts/%d", id))
}

func (c *Client) GetPostBySlug(slug string) (Post, error) {
	return get[Post](c, "/api/posts/slug/"+url.PathEscape(slug))
}

// CreatePost creates a new post and returns the created post and any CSS warnings.
func (c *Client) CreatePost(req CreatePostRequest) (Post, []string, error) {
	req.Content = strings.ReplaceAll(req.Content, "::: {", ":::{")
	r, err := post[postWriteResponse](c, "/api/posts", req)
	if err != nil {
		return Post{}, nil, err
	}
	return r.post(), r.CSSWarnings, nil
}

// UpdatePost replaces a post and returns the updated post and any CSS warnings.
func (c *Client) UpdatePost(id int64, req UpdatePostRequest) (Post, []string, error) {
	req.Content = strings.ReplaceAll(req.Content, "::: {", ":::{")
	r, err := put[postWriteResponse](c, fmt.Sprintf("/api/posts/%d", id), req)
	if err != nil {
		return Post{}, nil, err
	}
	return r.post(), r.CSSWarnings, nil
}

func (c *Client) Publish(id int64) (Post, error) {
	return post[Post](c, fmt.Sprintf("/api/posts/%d/publish", id), nil)
}

func (c *Client) Withdraw(id int64) (Post, error) {
	return delete[Post](c, fmt.Sprintf("/api/posts/%d/publish", id))
}

func (c *Client) DeletePost(id int64) error {
	return c.noContent(http.MethodDelete, fmt.Sprintf("/api/posts/%d", id), nil)
}

func (c *Client) GeneratePreviewLink(id int64) (string, error) {
	r, err := post[previewLinkResponse](c, fmt.Sprintf("/api/posts/%d/preview", id), nil)
	if err != nil {
		return "", err
	}
	return r.URL, nil
}

func (c *Client) UpdateTags(id int64, tags []string) (Post, error) {
	return post[Post](c, fmt.Sprintf("/api/posts/%d/tags", id), map[string][]string{"tags": tags})
}

func (c *Client) GetPostAnalytics() (PostAnalytics, error) {
	return get[PostAnalytics](c, "/api/posts/analytics")
}
