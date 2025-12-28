package showdoc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client ShowDoc API客户端
type Client struct {
	BaseURL    string // 基础URL，例如 https://www.showdoc.cc/server/api/open
	APIKey     string // API密钥
	APIToken   string // API令牌
	HTTPClient *http.Client
}

// NewClient 创建新的ShowDoc客户端
func NewClient(baseURL, apiKey, apiToken string) *Client {
	return &Client{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		APIToken:   apiToken,
		HTTPClient: &http.Client{},
	}
}

// BaseResponse 通用响应结构
type BaseResponse struct {
	ErrorCode    int         `json:"error_code"`
	ErrorMessage string      `json:"error_message"`
	Data         interface{} `json:"data"`
}

// PageRequest 页面请求基础结构
type PageRequest struct {
	APIKey   string `json:"api_key"`
	APIToken string `json:"api_token"`
}

// CatalogRequest 目录请求基础结构
type CatalogRequest struct {
	APIKey   string `json:"api_key"`
	APIToken string `json:"api_token"`
}

// postRequest 发送POST请求
func (c *Client) postRequest(endpoint string, requestBody interface{}) (*BaseResponse, error) {
	url := fmt.Sprintf("%s/%s", c.BaseURL, endpoint)

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var result BaseResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}
