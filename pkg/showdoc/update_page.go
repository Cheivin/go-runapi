package showdoc

import (
	"encoding/json"
	"fmt"

	"github.com/cheivin/go-runapi/pkg/types"
)

// UpdatePageRequest 更新或创建页面请求
type UpdatePageRequest struct {
	APIKey      string  `json:"api_key"`
	APIToken    string  `json:"api_token"`
	PageTitle   string  `json:"page_title"`
	PageContent string  `json:"page_content"`
	CatName     string  `json:"cat_name,omitempty"`
	SNumber     int     `json:"s_number,omitempty"`
	ExtInfo     ExtInfo `json:"ext_info"`
}
type ExtInfo struct {
	PageType string `json:"page_type"`
	APIInfo  struct {
		Method string `json:"method"`
	}
}

// UpdatePageResponse 更新页面响应
type UpdatePageResponse struct {
	PageID string `json:"page_id"`
}

// UpdatePage 更新或创建页面
// 按标题写入页面内容，自动创建目录
func (c *Client) UpdatePage(pageTitle string, pageContent types.PageContentFull, catName string, sNumber int) (*UpdatePageResponse, error) {
	// 将PageContentFull转换为JSON字符串
	jsonContent, err := json.Marshal(pageContent)
	if err != nil {
		return nil, fmt.Errorf("序列化页面内容失败: %v", err)
	}
	req := UpdatePageRequest{
		APIKey:      c.APIKey,
		APIToken:    c.APIToken,
		PageTitle:   pageTitle,
		PageContent: string(jsonContent),
		//PageContent: html.EscapeString(string(jsonContent)),
		CatName: catName,
		SNumber: sNumber,
		ExtInfo: ExtInfo{PageType: "api", APIInfo: struct {
			Method string `json:"method"`
		}{Method: pageContent.Info.Method}},
	}
	//if req.PageTitle == "用户登录" {
	//	req.PageContent = "{\"page_title\":\"用户登录\",\"info\":{\"from\":\"runapi\",\"type\":\"api\",\"title\":\"用户登录\",\"description\":\"用户登录的接口\",\"method\":\"post\",\"url\":\"http://127.0.0.1:8080/api/login\",\"remark\":\"\",\"apiStatus\":\"0\"},\"request\":{\"params\":{\"mode\":\"json\",\"urlencoded\":[],\"formdata\":[],\"json\":\"{}\",\"jsonDesc\":[{\"name\":\"username\",\"value\":\"\",\"type\":\"string\",\"require\":\"true\",\"remark\":\"用户名\"},{\"name\":\"password\",\"value\":\"\",\"type\":\"string\",\"require\":\"true\",\"remark\":\"密码\"},{\"name\":\"totp\",\"value\":\"\",\"type\":\"string\",\"require\":\"false\",\"remark\":\"双因子认证码（可选）\"},{\"name\":\"remember\",\"value\":\"\",\"type\":\"object\",\"require\":\"false\",\"remark\":\"记住登录状态（可选）\"},{\"name\":\"deviceId\",\"value\":\"\",\"type\":\"string\",\"require\":\"true\",\"remark\":\"设备ID（必传）\"}]},\"headers\":[],\"cookies\":[{\"name\":\"\",\"value\":\"\"}],\"auth\":{\"type\":\"none\",\"disabled\":\"0\"},\"query\":[],\"pathVariable\":[]},\"response\":{\"responseText\":\"{\\n  \\\"code\\\": 401,\\n  \\\"msg\\\": \\\"用户名或密码错误\\\",\\n  \\\"data\\\": null\\n}\",\"responseOriginal\":{\"code\":401,\"msg\":\"用户名或密码错误\",\"data\":null},\"responseExample\":\"\",\"responseHeader\":{\"content-length\":\"58\",\"content-type\":\"application/json\",\"cookie-from-server\":\"\",\"date\":\"Sun, 28 Dec 2025 09:34:34 GMT\"},\"responseStatus\":200,\"responseTime\":7,\"responseParamsDesc\":[{\"name\":\"code\",\"type\":\"int\",\"remark\":\"状态码\"},{\"name\":\"msg\",\"type\":\"string\",\"remark\":\"提示信息\"},{\"name\":\"data\",\"type\":\"object\",\"remark\":\"数据\"},{\"name\":\"data.user.id\",\"type\":\"int\",\"remark\":\"用户ID\"},{\"name\":\"data.user.username\",\"type\":\"string\",\"remark\":\"用户名\"},{\"name\":\"data.token\",\"type\":\"string\",\"remark\":\"登录凭证\"}],\"responseFailExample\":\"\",\"responseFailParamsDesc\":[{\"name\":\"\",\"type\":\"string\",\"remark\":\"\"}],\"remark\":\"登录接口\",\"responseSize\":0},\"scripts\":{\"pre\":\"\",\"post\":\"\"},\"testCases\":[],\"extend\":{}}"
	//}

	// 如果没有提供排序号，使用默认值99
	if sNumber == 0 {
		req.SNumber = 99
	}

	resp, err := c.postRequest("updatePage", req)
	if err != nil {
		return nil, err
	}

	if resp.ErrorCode != 0 {
		if resp.ErrorCode == 10101 {
			return nil, fmt.Errorf("API错误: %d - %s", resp.ErrorCode, "内容无变更")
		}
		return nil, fmt.Errorf("API错误: %d - %s", resp.ErrorCode, resp.ErrorMessage)
	}

	// 将Data转换为UpdatePageResponse
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("转换响应数据失败: %w", err)
	}

	var updateResp UpdatePageResponse
	if err := json.Unmarshal(dataBytes, &updateResp); err != nil {
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	return &updateResp, nil
}
