package point

import "fmt"

type tagListResponse struct {
	Tags []TagDetail `json:"tags"`
}

func (c *Client) ListTags() ([]TagDetail, error) {
	res, err := get[tagListResponse](c, "/api/tags")
	if err != nil {
		return nil, err
	}
	return res.Tags, nil
}

func (c *Client) CreateTag(req CreateTagRequest) (TagDetail, error) {
	return post[TagDetail](c, "/api/tags", req)
}

func (c *Client) GetTag(id int64) (TagDetail, error) {
	return get[TagDetail](c, fmt.Sprintf("/api/tags/%d", id))
}

func (c *Client) GetTagBySlug(slug string) (TagDetail, error) {
	return get[TagDetail](c, "/api/tags/slug/"+slug)
}

func (c *Client) UpdateTag(id int64, req UpdateTagRequest) (TagDetail, error) {
	return put[TagDetail](c, fmt.Sprintf("/api/tags/%d", id), req)
}

func (c *Client) DeleteTag(id int64) error {
	return c.noContent("DELETE", fmt.Sprintf("/api/tags/%d", id), nil)
}

func (c *Client) GeocodeTag(id int64) (TagLocation, error) {
	return post[TagLocation](c, fmt.Sprintf("/api/tags/%d/geocode", id), nil)
}
