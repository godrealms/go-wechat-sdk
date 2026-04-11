package oplatform

import "context"

// GetCategory 获取当前账号的类目。
// /cgi-bin/wxopen/getcategory
func (w *WxaAdminClient) GetCategory(ctx context.Context) (*WxaGetCategoryResp, error) {
	var resp WxaGetCategoryResp
	if err := w.doPost(ctx, "/cgi-bin/wxopen/getcategory", struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAllCategories 获取所有类目。
// /cgi-bin/wxopen/getallcategories
func (w *WxaAdminClient) GetAllCategories(ctx context.Context) (*WxaGetAllCategoriesResp, error) {
	var resp WxaGetAllCategoriesResp
	if err := w.doPost(ctx, "/cgi-bin/wxopen/getallcategories", struct{}{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddCategory 添加类目。
// /cgi-bin/wxopen/addcategory
func (w *WxaAdminClient) AddCategory(ctx context.Context, req *WxaAddCategoryReq) error {
	return w.doPost(ctx, "/cgi-bin/wxopen/addcategory", req, nil)
}

// DeleteCategory 删除类目。
// /cgi-bin/wxopen/deletecategory
func (w *WxaAdminClient) DeleteCategory(ctx context.Context, first, second int) error {
	body := map[string]int{"first": first, "second": second}
	return w.doPost(ctx, "/cgi-bin/wxopen/deletecategory", body, nil)
}

// ModifyCategory 修改类目资质。
// /cgi-bin/wxopen/modifycategory
func (w *WxaAdminClient) ModifyCategory(ctx context.Context, req *WxaModifyCategoryReq) error {
	return w.doPost(ctx, "/cgi-bin/wxopen/modifycategory", req, nil)
}
