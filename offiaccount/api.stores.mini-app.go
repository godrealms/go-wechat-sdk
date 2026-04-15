package offiaccount

import (
	"context"
	"fmt"
	"net/url"
)

// GetWxaStoreCateList 拉取门店小程序类目
func (c *Client) GetWxaStoreCateList(ctx context.Context) (*GetWxaStoreCateListResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := "/wxa/get_merchant_category"
	params := url.Values{"access_token": {token}}

	var result GetWxaStoreCateListResult
	if err := c.Https.Get(ctx, path, params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ApplyWxaStore 创建门店小程序
// req: 创建门店小程序请求参数
func (c *Client) ApplyWxaStore(ctx context.Context, req *ApplyWxaStoreRequest) (*ApplyWxaStoreResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/wxa/apply_merchant?access_token=%s", token)

	var result ApplyWxaStoreResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetWxaStoreAuditInfo 查询门店小程序审核结果
func (c *Client) GetWxaStoreAuditInfo(ctx context.Context) (*GetWxaStoreAuditInfoResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := "/wxa/get_merchant_audit_info"
	params := url.Values{"access_token": {token}}

	var result GetWxaStoreAuditInfoResult
	if err := c.Https.Get(ctx, path, params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// ModifyWxaStore 修改门店小程序信息
// req: 修改门店小程序信息请求参数
func (c *Client) ModifyWxaStore(ctx context.Context, req *ModifyWxaStoreRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/wxa/modify_merchant?access_token=%s", token)

	var result Resp
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDistrictList 获取省市区信息
func (c *Client) GetDistrictList(ctx context.Context) (*GetDistrictListResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := "/wxa/get_district"
	params := url.Values{"access_token": {token}}

	var result GetDistrictListResult
	if err := c.Https.Get(ctx, path, params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SearchMapPoi 搜索门店地图信息
// req: 搜索门店地图信息请求参数
func (c *Client) SearchMapPoi(ctx context.Context, req *SearchMapPoiRequest) (*SearchMapPoiResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/wxa/search_map_poi?access_token=%s", token)

	var result SearchMapPoiResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// AddStore 新增门店
// req: 新增门店请求参数
func (c *Client) AddStore(ctx context.Context, req *AddStoreRequest) (*AddStoreResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/wxa/add_store?access_token=%s", token)

	var result AddStoreResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStoreInfo 查询门店详情
// req: 查询门店详情请求参数
func (c *Client) GetStoreInfo(ctx context.Context, req *GetStoreInfoRequest) (*GetStoreInfoResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/wxa/get_store_info?access_token=%s", token)

	var result GetStoreInfoResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStoreList 查询门店列表
// req: 查询门店列表请求参数
func (c *Client) GetStoreList(ctx context.Context, req *GetStoreListRequest) (*GetStoreListResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/wxa/get_store_list?access_token=%s", token)

	var result GetStoreListResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DelStore 删除门店
// req: 删除门店请求参数
func (c *Client) DelStore(ctx context.Context, req *DelStoreRequest) (*Resp, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/wxa/del_store?access_token=%s", token)

	var result Resp
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateStore 更新门店信息
// req: 更新门店信息请求参数
func (c *Client) UpdateStore(ctx context.Context, req *UpdateStoreRequest) (*UpdateStoreResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/wxa/update_store?access_token=%s", token)

	var result UpdateStoreResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateMapPoi 在地图中创建门店
// req: 在地图中创建门店请求参数
func (c *Client) CreateMapPoi(ctx context.Context, req *CreateMapPoiRequest) (*CreateMapPoiResult, error) {
	token, err := c.AccessTokenE(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/wxa/create_map_poi?access_token=%s", token)

	var result CreateMapPoiResult
	if err := c.Https.Post(ctx, path, req, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
