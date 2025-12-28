package showdoc

import (
	"encoding/json"
	"fmt"
	"html"
)

// GetPageRequest 获取页面详情请求
type GetPageRequest struct {
	APIKey    string `json:"api_key"`
	APIToken  string `json:"api_token"`
	PageID    string `json:"page_id,omitempty"`
	PageTitle string `json:"page_title,omitempty"`
}

// PageData 页面数据结构
type PageData struct {
	PageID      string `json:"page_id"`
	PageTitle   string `json:"page_title"`
	PageContent string `json:"page_content"`
}

// GetPage 获取页面详情
// 通过page_id或page_title获取页面信息，二选一
func (c *Client) GetPage(pageID string, pageTitle string) (*PageData, error) {
	req := GetPageRequest{
		APIKey:    c.APIKey,
		APIToken:  c.APIToken,
		PageID:    pageID,
		PageTitle: pageTitle,
	}

	resp, err := c.postRequest("getPage", req)
	if err != nil {
		return nil, err
	}

	if resp.ErrorCode != 0 {
		return nil, fmt.Errorf("API错误: %d - %s", resp.ErrorCode, resp.ErrorMessage)
	}

	// 将Data转换为PageData
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("转换页面数据失败: %w", err)
	}

	var pageData PageData
	if err := json.Unmarshal(dataBytes, &pageData); err != nil {
		return nil, fmt.Errorf("解析页面数据失败: %w", err)
	}
	pageData.PageContent = html.UnescapeString(pageData.PageContent)

	return &pageData, nil
}

// GetPageByTitle 通过页面标题获取页面详情
func (c *Client) GetPageByTitle(pageTitle string) (*PageData, error) {
	return c.GetPage("", pageTitle)
}
