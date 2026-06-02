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

func (c *Client) CreateTag(name string) (TagDetail, error) {
	return post[TagDetail](c, "/api/tags", map[string]string{"name": name})
}
