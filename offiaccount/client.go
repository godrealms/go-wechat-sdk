package offiaccount

// Client 微信公众号
type Client struct {
	Config *Config
}

// NewClient 创建客户端
func NewClient(config *Config) *Client {
	return &Client{
		Config: config,
	}
}
