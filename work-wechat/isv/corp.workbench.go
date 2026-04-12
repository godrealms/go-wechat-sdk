package isv

import "context"

// SetWorkbenchTemplate 设置应用在工作台的展示模板。
func (cc *CorpClient) SetWorkbenchTemplate(ctx context.Context, req *WorkbenchTemplateReq) error {
	return cc.doPost(ctx, "/cgi-bin/agent/set_workbench_template", req, nil)
}

// GetWorkbenchTemplate 获取应用在工作台的展示模板。
// 注意：企业微信此接口也是 POST（不是 GET）。
func (cc *CorpClient) GetWorkbenchTemplate(ctx context.Context, agentID int) (*WorkbenchTemplateResp, error) {
	body := map[string]int{"agentid": agentID}
	var resp WorkbenchTemplateResp
	if err := cc.doPost(ctx, "/cgi-bin/agent/get_workbench_template", body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetWorkbenchData 设置指定用户在工作台上的个性化展示数据。
func (cc *CorpClient) SetWorkbenchData(ctx context.Context, req *WorkbenchDataReq) error {
	return cc.doPost(ctx, "/cgi-bin/agent/set_workbench_data", req, nil)
}
