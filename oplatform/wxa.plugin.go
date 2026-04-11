package oplatform

import "context"

// ApplyPlugin 使用方申请使用插件。
// POST /wxa/plugin body: {"action":"apply","plugin_appid":"..."}
func (w *WxaAdminClient) ApplyPlugin(ctx context.Context, pluginAppID string) error {
	body := map[string]string{
		"action":       "apply",
		"plugin_appid": pluginAppID,
	}
	return w.doPost(ctx, "/wxa/plugin", body, nil)
}

// ListPlugins 使用方查询已添加的插件列表。
// POST /wxa/plugin body: {"action":"list"}
func (w *WxaAdminClient) ListPlugins(ctx context.Context) (*WxaPluginList, error) {
	body := map[string]string{"action": "list"}
	var resp WxaPluginList
	if err := w.doPost(ctx, "/wxa/plugin", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UnbindPlugin 使用方解除插件。
// POST /wxa/plugin body: {"action":"unbind","plugin_appid":"..."}
func (w *WxaAdminClient) UnbindPlugin(ctx context.Context, pluginAppID string) error {
	body := map[string]string{
		"action":       "unbind",
		"plugin_appid": pluginAppID,
	}
	return w.doPost(ctx, "/wxa/plugin", body, nil)
}

// GetPluginDevApplyList 插件方：查询当前所有插件使用方申请列表。
// POST /wxa/devplugin body: {"action":"dev_apply_list","page":0,"num":10}
func (w *WxaAdminClient) GetPluginDevApplyList(ctx context.Context, page, num int) (*WxaPluginDevApplyList, error) {
	body := map[string]any{
		"action": "dev_apply_list",
		"page":   page,
		"num":    num,
	}
	var resp WxaPluginDevApplyList
	if err := w.doPost(ctx, "/wxa/devplugin", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AgreeDevPlugin 插件方：同意某个使用方的申请。
// POST /wxa/devplugin body: {"action":"dev_agree","appid":"..."}
func (w *WxaAdminClient) AgreeDevPlugin(ctx context.Context, userAppID string) error {
	body := map[string]string{
		"action": "dev_agree",
		"appid":  userAppID,
	}
	return w.doPost(ctx, "/wxa/devplugin", body, nil)
}

// RefuseDevPlugin 插件方：拒绝申请并给出原因。
// POST /wxa/devplugin body: {"action":"dev_refuse","reason":"..."}
func (w *WxaAdminClient) RefuseDevPlugin(ctx context.Context, reason string) error {
	body := map[string]string{
		"action": "dev_refuse",
		"reason": reason,
	}
	return w.doPost(ctx, "/wxa/devplugin", body, nil)
}

// DeleteDevPlugin 插件方：删除某个使用方的授权。
// POST /wxa/devplugin body: {"action":"dev_delete","appid":"..."}
func (w *WxaAdminClient) DeleteDevPlugin(ctx context.Context, userAppID string) error {
	body := map[string]string{
		"action": "dev_delete",
		"appid":  userAppID,
	}
	return w.doPost(ctx, "/wxa/devplugin", body, nil)
}
