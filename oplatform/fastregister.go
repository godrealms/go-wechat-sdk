package oplatform

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// FastRegisterClient 提供开放平台代注册小程序相关的 component 级接口。
//
// 所有方法都以 Client.ComponentAccessToken() 作为 token 源（而不是 authorizer
// 级别的 token），因为快速注册流程发生在 authorizer 关系建立之前。
//
// FastRegisterClient 无状态，线程安全，可在多 goroutine 共享。
type FastRegisterClient struct {
	c *Client
}

// FastRegister 从 Client 构造 FastRegisterClient。构造不做 I/O。
func (c *Client) FastRegister() *FastRegisterClient {
	return &FastRegisterClient{c: c}
}

// doPost 通用 POST 辅助：
//   - 从 Client 取 component_access_token 并以 query 参数形式拼接
//   - 正确处理 path 中已经带有 ?action=xxx 的情况（使用 & 分隔而非 ?）
//   - 复用包级 decodeRaw 进行两段式 JSON 解码（errcode 折叠 + typed unmarshal）
func (f *FastRegisterClient) doPost(ctx context.Context, path string, body, out any) error {
	ctx = touchContext(ctx)
	token, err := f.c.ComponentAccessToken(ctx)
	if err != nil {
		return err
	}
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	fullPath := path + sep + "component_access_token=" + url.QueryEscape(token)

	var raw json.RawMessage
	if err := f.c.http.Post(ctx, fullPath, body, &raw); err != nil {
		return fmt.Errorf("oplatform: %s: %w", path, err)
	}
	return decodeRaw(path, raw, out)
}

// CreateEnterpriseAccount 企业快速注册小程序。
// /cgi-bin/component/fastregisterweapp?action=create
func (f *FastRegisterClient) CreateEnterpriseAccount(ctx context.Context, req *FastRegEnterpriseReq) (*FastRegEnterpriseResp, error) {
	var resp FastRegEnterpriseResp
	if err := f.doPost(ctx, "/cgi-bin/component/fastregisterweapp?action=create", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryEnterpriseAccount 查询企业快速注册任务状态。
// /cgi-bin/component/fastregisterweapp?action=search
func (f *FastRegisterClient) QueryEnterpriseAccount(ctx context.Context, legalPersonaWechat, legalPersonaName string) (*FastRegEnterpriseStatus, error) {
	if legalPersonaWechat == "" {
		return nil, fmt.Errorf("oplatform: QueryEnterpriseAccount: legalPersonaWechat is required")
	}
	if legalPersonaName == "" {
		return nil, fmt.Errorf("oplatform: QueryEnterpriseAccount: legalPersonaName is required")
	}
	body := map[string]string{
		"legal_persona_wechat": legalPersonaWechat,
		"legal_persona_name":   legalPersonaName,
	}
	var resp FastRegEnterpriseStatus
	if err := f.doPost(ctx, "/cgi-bin/component/fastregisterweapp?action=search", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePersonalAccount 个人类型小程序快速注册。
// /cgi-bin/component/fastregisterpersonalweapp?action=create
func (f *FastRegisterClient) CreatePersonalAccount(ctx context.Context, req *FastRegPersonalReq) (*FastRegPersonalResp, error) {
	var resp FastRegPersonalResp
	if err := f.doPost(ctx, "/cgi-bin/component/fastregisterpersonalweapp?action=create", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryPersonalAccount 查询个人注册任务状态。
// /cgi-bin/component/fastregisterpersonalweapp?action=query
func (f *FastRegisterClient) QueryPersonalAccount(ctx context.Context, taskID string) (*FastRegPersonalStatus, error) {
	if taskID == "" {
		return nil, fmt.Errorf("oplatform: QueryPersonalAccount: taskID is required")
	}
	body := map[string]string{"taskid": taskID}
	var resp FastRegPersonalStatus
	if err := f.doPost(ctx, "/cgi-bin/component/fastregisterpersonalweapp?action=query", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateBetaAccount 复用主体创建试用版小程序。
// /cgi-bin/component/fastregisterbetaweapp?action=create
func (f *FastRegisterClient) CreateBetaAccount(ctx context.Context, req *FastRegBetaReq) (*FastRegBetaResp, error) {
	var resp FastRegBetaResp
	if err := f.doPost(ctx, "/cgi-bin/component/fastregisterbetaweapp?action=create", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryBetaAccount 查询试用版创建任务状态。
// /cgi-bin/component/fastregisterbetaweapp?action=search
func (f *FastRegisterClient) QueryBetaAccount(ctx context.Context, uniqueID string) (*FastRegBetaStatus, error) {
	if uniqueID == "" {
		return nil, fmt.Errorf("oplatform: QueryBetaAccount: uniqueID is required")
	}
	body := map[string]string{"unique_id": uniqueID}
	var resp FastRegBetaStatus
	if err := f.doPost(ctx, "/cgi-bin/component/fastregisterbetaweapp?action=search", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GenerateAdminRebindQrcode 生成小程序管理员变更二维码。
// /cgi-bin/account/componentrebindadmin
func (f *FastRegisterClient) GenerateAdminRebindQrcode(ctx context.Context, redirectURI string) (*RebindAdminQrcode, error) {
	body := map[string]string{"redirect_uri": redirectURI}
	var resp RebindAdminQrcode
	if err := f.doPost(ctx, "/cgi-bin/account/componentrebindadmin", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
