package point

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

func (c *Client) GetTag(slug string) (TagDetail, error) {
	return get[TagDetail](c, "/api/tags/"+slug)
}

func (c *Client) UpdateTag(slug string, req UpdateTagRequest) (TagDetail, error) {
	return put[TagDetail](c, "/api/tags/"+slug, req)
}

func (c *Client) DeleteTag(slug string) error {
	return c.noContent("DELETE", "/api/tags/"+slug, nil)
}
