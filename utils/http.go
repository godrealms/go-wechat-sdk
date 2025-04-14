package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTP 客户端结构体
type HTTP struct {
	BaseURL string
	Client  *http.Client
	Headers map[string]string
	Timeout time.Duration
}

// Option 定义配置选项的函数类型
type Option func(*HTTP)

// NewHTTP 创建新的 HTTP 客户端
func NewHTTP(baseURL string, opts ...Option) *HTTP {
	h := &HTTP{
		BaseURL: baseURL,
		Headers: make(map[string]string),
		Timeout: 30 * time.Second,
	}

	// 应用所有选项
	for _, opt := range opts {
		opt(h)
	}

	// 初始化 http.Client
	h.Client = &http.Client{
		Timeout: h.Timeout,
	}

	return h
}

// WithTimeout 设置超时时间的选项
func WithTimeout(timeout time.Duration) Option {
	return func(h *HTTP) {
		h.Timeout = timeout
	}
}

// WithHeaders 设置请求头的选项
func WithHeaders(headers map[string]string) Option {
	return func(h *HTTP) {
		for k, v := range headers {
			h.Headers[k] = v
		}
	}
}

// SetBaseURL 设置基础URL
func (h *HTTP) SetBaseURL(url string) {
	h.BaseURL = url
}

// do 执行 HTTP 请求的通用方法
func (h *HTTP) do(ctx context.Context, method, path string, body interface{}, query url.Values, result interface{}) error {
	if query != nil {
		path += "?" + query.Encode()
	}
	// 构建完整URL
	fullURL := h.BaseURL + path

	log.Println("method:", method)
	log.Println("fullURL:", fullURL)

	// 处理请求体
	var bodyReader io.Reader
	var bodyJson string
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
		bodyJson = string(jsonBody)
		log.Println("body:", bodyJson)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置默认请求头
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 添加自定义请求头
	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := h.Client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	// 如果需要解析响应结果
	if result != nil && len(respBody) > 0 {
		if err = json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response body failed: %w: %s", err, string(respBody))
		}
	}

	return nil
}

// Get 发送 GET 请求
func (h *HTTP) Get(ctx context.Context, path string, query url.Values, result interface{}) error {
	return h.do(ctx, http.MethodGet, path, nil, query, result)
}

// Post 发送 POST 请求
func (h *HTTP) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return h.do(ctx, http.MethodPost, path, body, nil, result)
}

// Put 发送 PUT 请求
func (h *HTTP) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return h.do(ctx, http.MethodPut, path, body, nil, result)
}

// Patch 发送 PATCH 请求
func (h *HTTP) Patch(ctx context.Context, path string, body interface{}, result interface{}) error {
	return h.do(ctx, http.MethodPatch, path, body, nil, result)
}

// Delete 发送 DELETE 请求
func (h *HTTP) Delete(ctx context.Context, path string, result interface{}) error {
	return h.do(ctx, http.MethodDelete, path, nil, nil, result)
}

// PostForm 发送 POST 表单请求
func (h *HTTP) PostForm(ctx context.Context, path string, form url.Values, result interface{}) error {
	fullURL := h.BaseURL + path

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	// 设置表单请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 添加自定义请求头
	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := h.Client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应结果
	if result != nil && len(respBody) > 0 {
		if err = json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response body failed: %w:%s", err, string(respBody))
		}
	}

	return nil
}
