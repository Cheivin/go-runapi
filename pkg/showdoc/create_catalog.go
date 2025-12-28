package showdoc

import (
	"encoding/json"
	"fmt"
)

// CreateCatalogRequest 创建目录请求
type CreateCatalogRequest struct {
	APIKey      string `json:"api_key"`
	APIToken    string `json:"api_token"`
	CatName     string `json:"cat_name"`
	ParentCatID string `json:"parent_cat_id,omitempty"`
	SNumber     string `json:"s_number,omitempty"`
}

// CreateCatalogResponse 创建目录响应
type CreateCatalogResponse struct {
	CatID string `json:"cat_id"`
}

// CreateCatalog 创建目录
// 新建目录节点，可指定父级
func (c *Client) CreateCatalog(catName, parentCatID, sNumber string) (*CreateCatalogResponse, error) {
	req := CreateCatalogRequest{
		APIKey:      c.APIKey,
		APIToken:    c.APIToken,
		CatName:     catName,
		ParentCatID: parentCatID,
		SNumber:     sNumber,
	}

	// 如果没有提供排序号，使用默认值99
	if sNumber == "0" {
		req.SNumber = "99"
	}

	resp, err := c.postRequest("createCatalog", req)
	if err != nil {
		return nil, err
	}

	if resp.ErrorCode != 0 {
		return nil, fmt.Errorf("API错误: %d - %s", resp.ErrorCode, resp.ErrorMessage)
	}

	// 将Data转换为CreateCatalogResponse
	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("转换响应数据失败: %w", err)
	}

	var createResp CreateCatalogResponse
	if err := json.Unmarshal(dataBytes, &createResp); err != nil {
		return nil, fmt.Errorf("解析响应数据失败: %w", err)
	}

	return &createResp, nil
}

// CreateRootCatalog 创建根目录
func (c *Client) CreateRootCatalog(catName string) (*CreateCatalogResponse, error) {
	return c.CreateCatalog(catName, "0", "99")
}

// CreateSubCatalog 创建子目录
func (c *Client) CreateSubCatalog(catName string, parentCatID string) (*CreateCatalogResponse, error) {
	return c.CreateCatalog(catName, parentCatID, "99")
}

// CreateCatalogWithOrder 创建带排序的目录
func (c *Client) CreateCatalogWithOrder(catName, parentCatID, sNumber string) (*CreateCatalogResponse, error) {
	return c.CreateCatalog(catName, parentCatID, sNumber)
}
