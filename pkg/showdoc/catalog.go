package showdoc

import (
	"encoding/json"
	"fmt"
)

// GetCatalogTreeRequest 获取目录树请求
type GetCatalogTreeRequest struct {
	APIKey   string `json:"api_key"`
	APIToken string `json:"api_token"`
}

// CatalogItem 目录项结构
type CatalogItem struct {
	CatID    string        `json:"cat_id"`
	CatName  string        `json:"cat_name"`
	Catalogs []CatalogItem `json:"catalogs,omitempty"`
	Level    string        `json:"level"`
	Pages    []PageItem    `json:"pages,omitempty"`
	ParentID string        `json:"parent_cat_id"`
	SNumber  string        `json:"s_number"`
}

// PageItem 页面项结构
type PageItem struct {
	CatID     string `json:"cat_id"`
	PageID    string `json:"page_id"`
	PageTitle string `json:"page_title"`
	SNumber   string `json:"s_number"`
}

// CatalogTreeData 目录树数据结构
type CatalogTreeData struct {
	Pages    []PageItem    `json:"pages"`
	Catalogs []CatalogItem `json:"catalogs"`
}

// GetCatalogTree 获取目录树
// 返回项目完整的目录及页面树结构
func (c *Client) GetCatalogTree() (*CatalogTreeData, error) {
	req := GetCatalogTreeRequest{
		APIKey:   c.APIKey,
		APIToken: c.APIToken,
	}

	resp, err := c.postRequest("getCatalogTree", req)
	if err != nil {
		return nil, err
	}

	if resp.ErrorCode != 0 {
		return nil, fmt.Errorf("API错误: %d - %s", resp.ErrorCode, resp.ErrorMessage)
	}

	// 将Data转换为CatalogTreeData
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("转换目录树数据失败: %w", err)
	}

	var treeData CatalogTreeData
	if err := json.Unmarshal(dataBytes, &treeData); err != nil {
		return nil, fmt.Errorf("解析目录树数据失败: %w", err)
	}

	return &treeData, nil
}

// GetCatalogs 获取目录列表
func (c *Client) GetCatalogs() ([]CatalogItem, error) {
	treeData, err := c.GetCatalogTree()
	if err != nil {
		return nil, err
	}
	return treeData.Catalogs, nil
}

// GetPages 获取页面列表
func (c *Client) GetPages() ([]PageItem, error) {
	treeData, err := c.GetCatalogTree()
	if err != nil {
		return nil, err
	}
	return treeData.Pages, nil
}
