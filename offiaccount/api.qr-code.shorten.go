package offiaccount

// GenShortKey 长信息转短链
// longData: 需要转换的长信息，不超过4KB
// expireSeconds: 过期秒数，最大值为2592000（即30天），默认为2592000
func (c *Client) GenShortKey(longData string, expireSeconds int64) (*GenShortKeyResult, error) {
	// 构造请求URL
	path := "/cgi-bin/shorten/gen"

	// 构造请求体
	req := &GenShortKeyRequest{
		LongData:      longData,
		ExpireSeconds: expireSeconds,
	}

	// 发送请求
	var result GenShortKeyResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// FetchShorten 短链转长信息
// shortKey: 短key
func (c *Client) FetchShorten(shortKey string) (*FetchShortenResult, error) {
	// 构造请求URL
	path := "/cgi-bin/shorten/fetch"

	// 构造请求体
	req := &FetchShortenRequest{
		ShortKey: shortKey,
	}

	// 发送请求
	var result FetchShortenResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
