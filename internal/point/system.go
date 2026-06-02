package point

func (c *Client) GetStats() (Stats, error) {
	return get[Stats](c, "/api/system/stats")
}
