package offiaccount

import (
	"context"
	"fmt"
)

// GenShortKey 长信息转短链
// longData: 需要转换的长信息，不超过4KB
// expireSeconds: 过期秒数，最大值为2592000（即30天），默认为2592000
func (c *Client) GenShortKey(ctx context.Context, longData string, expireSeconds int64) (*GenShortKeyResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/shorten/gen?access_token=%s", token)

	req := &GenShortKeyRequest{
		LongData:      longData,
		ExpireSeconds: expireSeconds,
	}

	var result GenShortKeyResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// FetchShorten 短链转长信息
// shortKey: 短key
func (c *Client) FetchShorten(ctx context.Context, shortKey string) (*FetchShortenResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/cgi-bin/shorten/fetch?access_token=%s", token)

	req := &FetchShortenRequest{
		ShortKey: shortKey,
	}

	var result FetchShortenResult
	if err := c.doPost(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
