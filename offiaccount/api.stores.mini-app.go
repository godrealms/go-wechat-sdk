package offiaccount

// GetWxaStoreCateList 拉取门店小程序类目
func (c *Client) GetWxaStoreCateList() (*GetWxaStoreCateListResult, error) {
	// 构造请求URL
	path := "/wxa/get_merchant_category"

	// 发送请求
	var result GetWxaStoreCateListResult
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ApplyWxaStore 创建门店小程序
// req: 创建门店小程序请求参数
func (c *Client) ApplyWxaStore(req *ApplyWxaStoreRequest) (*ApplyWxaStoreResult, error) {
	// 构造请求URL
	path := "/wxa/apply_merchant"

	// 发送请求
	var result ApplyWxaStoreResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetWxaStoreAuditInfo 查询门店小程序审核结果
func (c *Client) GetWxaStoreAuditInfo() (*GetWxaStoreAuditInfoResult, error) {
	// 构造请求URL
	path := "/wxa/get_merchant_audit_info"

	// 发送请求
	var result GetWxaStoreAuditInfoResult
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ModifyWxaStore 修改门店小程序信息
// req: 修改门店小程序信息请求参数
func (c *Client) ModifyWxaStore(req *ModifyWxaStoreRequest) (*Resp, error) {
	// 构造请求URL
	path := "/wxa/modify_merchant"

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetDistrictList 获取省市区信息
func (c *Client) GetDistrictList() (*GetDistrictListResult, error) {
	// 构造请求URL
	path := "/wxa/get_district"

	// 发送请求
	var result GetDistrictListResult
	err := c.Https.Get(c.ctx, path, nil, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SearchMapPoi 搜索门店地图信息
// req: 搜索门店地图信息请求参数
func (c *Client) SearchMapPoi(req *SearchMapPoiRequest) (*SearchMapPoiResult, error) {
	// 构造请求URL
	path := "/wxa/search_map_poi"

	// 发送请求
	var result SearchMapPoiResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// AddStore 新增门店
// req: 新增门店请求参数
func (c *Client) AddStore(req *AddStoreRequest) (*AddStoreResult, error) {
	// 构造请求URL
	path := "/wxa/add_store"

	// 发送请求
	var result AddStoreResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStoreInfo 查询门店详情
// req: 查询门店详情请求参数
func (c *Client) GetStoreInfo(req *GetStoreInfoRequest) (*GetStoreInfoResult, error) {
	// 构造请求URL
	path := "/wxa/get_store_info"

	// 发送请求
	var result GetStoreInfoResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStoreList 查询门店列表
// req: 查询门店列表请求参数
func (c *Client) GetStoreList(req *GetStoreListRequest) (*GetStoreListResult, error) {
	// 构造请求URL
	path := "/wxa/get_store_list"

	// 发送请求
	var result GetStoreListResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DelStore 删除门店
// req: 删除门店请求参数
func (c *Client) DelStore(req *DelStoreRequest) (*Resp, error) {
	// 构造请求URL
	path := "/wxa/del_store"

	// 发送请求
	var result Resp
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateStore 更新门店信息
// req: 更新门店信息请求参数
func (c *Client) UpdateStore(req *UpdateStoreRequest) (*UpdateStoreResult, error) {
	// 构造请求URL
	path := "/wxa/update_store"

	// 发送请求
	var result UpdateStoreResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CreateMapPoi 在地图中创建门店
// req: 在地图中创建门店请求参数
func (c *Client) CreateMapPoi(req *CreateMapPoiRequest) (*CreateMapPoiResult, error) {
	// 构造请求URL
	path := "/wxa/create_map_poi"

	// 发送请求
	var result CreateMapPoiResult
	err := c.Https.Post(c.ctx, path, req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
