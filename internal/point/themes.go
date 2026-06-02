package point

func (c *Client) ListThemes() ([]Theme, error) {
	return get[[]Theme](c, "/api/themes")
}

func (c *Client) GetActiveTheme() (Theme, error) {
	return get[Theme](c, "/api/themes/active")
}

func (c *Client) SetActiveTheme(name string) (Theme, error) {
	return put[Theme](c, "/api/themes/active", map[string]string{"name": name})
}
