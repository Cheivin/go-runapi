package showdoc

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/cheivin/go-runapi/pkg/config"
	"github.com/cheivin/go-runapi/pkg/generator"
	"github.com/cheivin/go-runapi/pkg/types"
)

// Pusher ShowDoc推送器
type Pusher struct {
	client *Client
	config *config.ShowDocConfig
}

// NewPusher 创建新的推送器
func NewPusher(cfg *config.ShowDocConfig) *Pusher {
	return &Pusher{
		client: NewClient(cfg.URL, cfg.APIKey, cfg.APIToken),
		config: cfg,
	}
}

// PushDocuments 推送文档到ShowDoc
func (p *Pusher) PushDocuments(docs []types.APIDoc) error {
	if !p.config.Enabled {
		return fmt.Errorf("ShowDoc推送未启用")
	}

	if p.config.APIKey == "" || p.config.APIToken == "" {
		return fmt.Errorf("ShowDoc API密钥或令牌未配置")
	}

	fmt.Printf("开始推送 %d 个API文档到ShowDoc...\n", len(docs))

	// 首先获取目录树，了解现有结构
	catalogTree, err := p.client.GetCatalogTree()
	if err != nil {
		return fmt.Errorf("获取ShowDoc目录树失败: %v", err)
	}

	// 创建目录映射
	catalogMap := p.buildCatalogMap(catalogTree.Catalogs)

	// 按目录分组文档
	categorizedDocs := p.categorizeDocuments(docs)

	// 推送每个目录的文档
	for catalogPath, catalogDocs := range categorizedDocs {
		if err := p.pushCatalogDocuments(catalogPath, catalogDocs, catalogMap); err != nil {
			return fmt.Errorf("推送目录 %s 的文档失败: %v", catalogPath, err)
		}
	}

	fmt.Printf("成功推送所有文档到ShowDoc\n")
	return nil
}

// PushChangedDocuments 只推送有变更的文档
func (p *Pusher) PushChangedDocuments(diff *generator.DocumentDiff) error {
	if !p.config.Enabled {
		return fmt.Errorf("ShowDoc推送未启用")
	}

	if p.config.APIKey == "" || p.config.APIToken == "" {
		return fmt.Errorf("ShowDoc API密钥或令牌未配置")
	}

	if !diff.HasChanges() {
		fmt.Println("没有文档变更，跳过推送")
		return nil
	}

	fmt.Printf("检测到文档变更: %s\n", diff.GetSummary())

	// 首先获取目录树
	catalogTree, err := p.client.GetCatalogTree()
	if err != nil {
		return fmt.Errorf("获取ShowDoc目录树失败: %v", err)
	}

	// 创建目录映射
	catalogMap := p.buildCatalogMap(catalogTree.Catalogs)

	// 处理新增的文档
	if len(diff.Added) > 0 {
		fmt.Printf("推送 %d 个新增文档...\n", len(diff.Added))
		categorizedDocs := p.categorizeDocuments(diff.Added)
		for catalogPath, catalogDocs := range categorizedDocs {
			if err := p.pushCatalogDocuments(catalogPath, catalogDocs, catalogMap); err != nil {
				return fmt.Errorf("推送新增文档失败: %v", err)
			}
		}
	}

	// 处理修改的文档
	if len(diff.Changed) > 0 {
		fmt.Printf("更新 %d 个修改文档...\n", len(diff.Changed))
		for _, change := range diff.Changed {
			if err := p.pushSingleDocument(change.New); err != nil {
				log.Printf("更新文档 %s 失败: %v", change.New.Title, err)
			}
		}
	}
	return nil
}

// buildCatalogMap 构建目录映射
func (p *Pusher) buildCatalogMap(catalogs []CatalogItem) map[string]string {
	catalogMap := make(map[string]string)
	p.buildCatalogMapRecursive(catalogs, "", catalogMap)
	return catalogMap
}

// buildCatalogMapRecursive 递归构建目录映射
func (p *Pusher) buildCatalogMapRecursive(catalogs []CatalogItem, parentPath string, catalogMap map[string]string) {
	for _, catalog := range catalogs {
		// 构建当前目录的完整路径
		var currentPath string
		if parentPath == "" {
			currentPath = catalog.CatName
		} else {
			currentPath = parentPath + "/" + catalog.CatName
		}

		// 添加到映射
		catalogMap[currentPath] = catalog.CatID

		// 递归处理子目录
		if len(catalog.Catalogs) > 0 {
			p.buildCatalogMapRecursive(catalog.Catalogs, currentPath, catalogMap)
		}
	}
}

// categorizeDocuments 按目录分组文档
func (p *Pusher) categorizeDocuments(docs []types.APIDoc) map[string][]types.APIDoc {
	categorized := make(map[string][]types.APIDoc)

	for _, doc := range docs {
		catalogPath := doc.Catalog
		if catalogPath == "" {
			catalogPath = "默认目录"
		}

		categorized[catalogPath] = append(categorized[catalogPath], doc)
	}

	return categorized
}

// pushCatalogDocuments 推送单个目录的文档
func (p *Pusher) pushCatalogDocuments(catalogPath string, docs []types.APIDoc, catalogMap map[string]string) error {
	// 确保目录存在
	_, err := p.ensureCatalogExists(catalogPath, catalogMap)
	if err != nil {
		return err
	}

	// 推送文档
	for _, doc := range docs {
		if err := p.pushSingleDocument(doc); err != nil {
			log.Printf("推送文档 %s 失败: %v", doc.Title, err)
		}
	}

	return nil
}

// ensureCatalogExists 确保目录存在
func (p *Pusher) ensureCatalogExists(catalogPath string, catalogMap map[string]string) (string, error) {
	// 如果目录已存在，直接返回ID
	if catalogID, exists := catalogMap[catalogPath]; exists {
		return catalogID, nil
	}

	// 创建目录层级
	parts := strings.Split(catalogPath, "/")
	parentID := "0"
	currentPath := ""

	for i, part := range parts {
		if i > 0 {
			currentPath += "/"
		}
		currentPath += part

		if catalogID, exists := catalogMap[currentPath]; exists {
			parentID = catalogID
			continue
		}

		// 创建新目录
		resp, err := p.client.CreateCatalog(part, parentID, "99")
		if err != nil {
			return "", fmt.Errorf("创建目录 %s 失败: %v", part, err)
		}

		fmt.Printf("创建目录: %s (ID: %s)\n", currentPath, resp.CatID)
		catalogMap[currentPath] = resp.CatID
		parentID = resp.CatID
	}

	return parentID, nil
}

// pushSingleDocument 推送单个文档
func (p *Pusher) pushSingleDocument(doc types.APIDoc) error {
	// 转换为ShowDoc页面内容结构
	pageContent := types.APIDocToPageContent(doc)

	// 获取现有页面
	existingPageContent, err := p.getExistingPageContent(doc.Title)
	if err != nil {
		fullContent := types.CreateDefaultFullContent()
		pageContentFull := types.MergeWithFullContent(pageContent, fullContent)
		fmt.Printf("无法获取现有页面，直接推送新页面: %s\n", doc.Title)
		return p.pushPageContent(doc, pageContentFull)
	}
	// 合并现有内容和新内容
	mergedContent := types.MergeWithFullContent(pageContent, *existingPageContent)
	// 推送合并后的内容
	return p.pushPageContent(doc, mergedContent)
}

// getExistingPageContent 获取现有页面内容
func (p *Pusher) getExistingPageContent(pageTitle string) (*types.PageContentFull, error) {
	// 尝试通过标题获取页面
	pageData, err := p.client.GetPageByTitle(pageTitle)
	if err != nil {
		return nil, fmt.Errorf("获取页面失败: %v", err)
	}
	fullContent := new(types.PageContentFull)
	err = json.Unmarshal([]byte(pageData.PageContent), fullContent)
	if err != nil {
		return nil, fmt.Errorf("解析页面内容失败: %v", err)
	}

	//// 更新基本信息
	fullContent.Info.Title = pageData.PageTitle
	return fullContent, nil
}

// mergePageContent 合并页面内容
func (p *Pusher) mergePageContent(existing, new types.PageContentFull) types.PageContentFull {
	// 使用新内容覆盖现有内容，但保留一些现有配置
	merged := existing

	// 更新基本信息
	merged.Info.Title = new.Info.Title
	merged.Info.Description = new.Info.Description
	merged.Info.Method = new.Info.Method
	merged.Info.URL = new.Info.URL

	// 更新请求信息
	merged.Request.Params.Mode = new.Request.Params.Mode
	merged.Request.Params.URLEncoded = new.Request.Params.URLEncoded
	merged.Request.Params.FormData = new.Request.Params.FormData
	merged.Request.Params.JSONDesc = new.Request.Params.JSONDesc
	merged.Request.Headers = new.Request.Headers
	merged.Request.Query = new.Request.Query

	// 更新响应信息
	merged.Response.ResponseParamsDesc = new.Response.ResponseParamsDesc
	merged.Response.Remark = new.Response.Remark

	// 保留现有的其他信息（如scripts、testCases等）

	return merged
}

// pushPageContent 推送页面内容
func (p *Pusher) pushPageContent(doc types.APIDoc, pageContent types.PageContentFull) error {
	// 推送到ShowDoc
	resp, err := p.client.UpdatePage(doc.Title, pageContent, doc.Catalog, 99)
	if err != nil {
		return fmt.Errorf("推送页面失败: %v", err)
	}

	fmt.Printf("推送文档: %s (页面ID: %s)\n", doc.Title, resp.PageID)
	return nil
}

// getRouter 获取路由信息
func (p *Pusher) getRouter(doc types.APIDoc) string {
	if doc.Router != "" {
		return doc.Router
	}
	if doc.URL != "" {
		return doc.URL
	}
	return ""
}
