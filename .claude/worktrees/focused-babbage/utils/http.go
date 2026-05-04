package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultMaxResponseSize 默认最大响应体大小 (10MB)
	DefaultMaxResponseSize = 10 * 1024 * 1024
)

// Logger 日志接口
type Logger interface {
	Printf(format string, v ...interface{})
}

// HTTP 客户端结构体
type HTTP struct {
	BaseURL string
	Client  *http.Client
	Headers map[string]string
	Timeout time.Duration
	Logger  Logger
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

// WithLogger 设置日志记录器
func WithLogger(logger Logger) Option {
	return func(h *HTTP) {
		h.Logger = logger
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

	if h.Logger != nil {
		h.Logger.Printf("method: %s", method)
		h.Logger.Printf("url: %s", fullURL)
	}

	// 处理请求体
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
		if h.Logger != nil {
			h.Logger.Printf("body: %s", string(jsonBody))
		}
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

	// 读取响应体（限制最大大小防止OOM）
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, DefaultMaxResponseSize))
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	if h.Logger != nil {
		h.Logger.Printf("response: %s", string(respBody))
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

// doWithHeaders 执行 HTTP 请求，使用请求级别的 headers（线程安全）
func (h *HTTP) doWithHeaders(ctx context.Context, method, path string, body interface{}, reqHeaders map[string]string, query url.Values, result interface{}) error {
	if query != nil {
		path += "?" + query.Encode()
	}
	fullURL := h.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}
	for k, v := range reqHeaders {
		req.Header.Set(k, v)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, DefaultMaxResponseSize))
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err = json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response body failed: %w: %s", err, string(respBody))
		}
	}
	return nil
}

// PostWithHeaders 发送 POST 请求，使用请求级别的 headers（线程安全）
func (h *HTTP) PostWithHeaders(ctx context.Context, path string, body interface{}, headers map[string]string, result interface{}) error {
	return h.doWithHeaders(ctx, http.MethodPost, path, body, headers, nil, result)
}

// GetWithHeaders 发送 GET 请求，使用请求级别的 headers（线程安全）
func (h *HTTP) GetWithHeaders(ctx context.Context, path string, headers map[string]string, query url.Values, result interface{}) error {
	return h.doWithHeaders(ctx, http.MethodGet, path, nil, headers, query, result)
}

// PostBinary 发送 POST 请求，返回原始字节（用于图片等二进制响应）
func (h *HTTP) PostBinary(ctx context.Context, path string, body interface{}) ([]byte, error) {
	fullURL := h.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body failed: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, DefaultMaxResponseSize))
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
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

	// 读取响应体（限制最大大小防止OOM）
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, DefaultMaxResponseSize))
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

// PostMultipart sends a multipart/form-data POST request with a single file field.
func (h *HTTP) PostMultipart(ctx context.Context, path string, fieldName string, fileName string, fileData []byte, result interface{}) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile(fieldName, fileName)
	if err != nil {
		return fmt.Errorf("create form file failed: %w", err)
	}
	if _, err = fw.Write(fileData); err != nil {
		return fmt.Errorf("write file data failed: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("close multipart writer failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.BaseURL+path, &buf)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := h.Client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, DefaultMaxResponseSize))
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}
	if result != nil && len(respBody) > 0 {
		if err = json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response body failed: %w: %s", err, string(respBody))
		}
	}
	return nil
}
