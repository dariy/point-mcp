package point

func (c *Client) GetPublicSettings() (Settings, error) {
	return get[Settings](c, "/api/settings/public")
}

func (c *Client) GetSettings() (Settings, error) {
	return get[Settings](c, "/api/settings")
}

func (c *Client) UpdateSettings(updates map[string]string) (Settings, error) {
	return patch[Settings](c, "/api/settings", updates)
}
