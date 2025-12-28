package generator

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cheivin/go-runapi/internal/parser"
	"github.com/cheivin/go-runapi/pkg/config"
	"github.com/cheivin/go-runapi/pkg/types"
)

// Generator 文档生成器
type Generator struct {
	parser *parser.Parser
	config *config.Config
}

// NewGenerator 创建新的文档生成器
func NewGenerator(cfg *config.Config) *Generator {
	// 构建所有扫描目录：根目录 + 额外目录
	allDirs := append([]string{cfg.Scan.Dir}, cfg.Scan.ExtraDirs...)

	return &Generator{
		parser: parser.NewParser(cfg.Scan.Scan, allDirs, cfg.Scan.IncludeVendor),
		config: cfg,
	}
}

// GenerateDocuments 生成文档
func (g *Generator) GenerateDocuments() (bool, error) {
	// 解析API文档
	apiDocs, err := g.parser.ParseDir()
	if err != nil {
		return false, fmt.Errorf("解析API文档失败: %v", err)
	}

	if len(apiDocs) == 0 {
		fmt.Println("未找到任何API文档")
		return false, nil
	}

	// 生成JSON内容
	jsonContent, err := g.parser.GenerateJSON(apiDocs)
	if err != nil {
		return false, fmt.Errorf("生成JSON文档失败: %v", err)
	}

	// 检查文件是否有变化
	changed, err := g.hasFileChanged(g.config.Output.File, jsonContent)
	if err != nil {
		return false, fmt.Errorf("检查文件变化失败: %v", err)
	}

	if !changed {
		fmt.Printf("文档文件 %s 无变化，跳过生成\n", g.config.Output.File)
		return false, nil
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(g.config.Output.File)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return false, fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(g.config.Output.File, []byte(jsonContent), 0644); err != nil {
		return false, fmt.Errorf("写入文档文件失败: %v", err)
	}

	fmt.Printf("文档已生成到: %s (%d个API)\n", g.config.Output.File, len(apiDocs))
	return true, nil
}

// hasFileChanged 检查文件内容是否有变化
func (g *Generator) hasFileChanged(filePath, newContent string) (bool, error) {
	// 如果文件不存在，认为有变化
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return true, nil
	}

	// 读取现有文件内容
	existingContent, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	// 比较内容的MD5值
	existingMD5 := md5.Sum(existingContent)
	newMD5 := md5.Sum([]byte(newContent))

	return existingMD5 != newMD5, nil
}

// GetGeneratedDocuments 获取生成的文档内容
func (g *Generator) GetGeneratedDocuments() ([]types.APIDoc, string, error) {
	// 解析API文档
	apiDocs, err := g.parser.ParseDir()
	if err != nil {
		return nil, "", fmt.Errorf("解析API文档失败: %v", err)
	}

	if len(apiDocs) == 0 {
		return nil, "", fmt.Errorf("未找到任何API文档")
	}

	// 生成JSON内容
	jsonContent, err := g.parser.GenerateJSON(apiDocs)
	if err != nil {
		return nil, "", fmt.Errorf("生成JSON文档失败: %v", err)
	}

	return apiDocs, jsonContent, nil
}

// LoadExistingDocuments 加载现有文档
func (g *Generator) LoadExistingDocuments() ([]types.APIDoc, error) {
	// 检查文件是否存在
	if _, err := os.Stat(g.config.Output.File); os.IsNotExist(err) {
		return nil, fmt.Errorf("文档文件不存在: %s", g.config.Output.File)
	}

	// 读取文件内容
	data, err := os.ReadFile(g.config.Output.File)
	if err != nil {
		return nil, fmt.Errorf("读取文档文件失败: %v", err)
	}

	// 解析JSON
	var apiDocs []types.APIDoc
	if err := json.Unmarshal(data, &apiDocs); err != nil {
		return nil, fmt.Errorf("解析文档文件失败: %v", err)
	}

	return apiDocs, nil
}

// CompareDocuments 比较两个文档列表的差异
func (g *Generator) CompareDocuments(oldDocs, newDocs []types.APIDoc) *DocumentDiff {
	diff := &DocumentDiff{
		Added:   []types.APIDoc{},
		Removed: []types.APIDoc{},
		Changed: []DocumentChange{},
	}

	// 创建旧文档的映射
	oldDocMap := make(map[string]types.APIDoc)
	for _, doc := range oldDocs {
		key := g.getDocKey(doc)
		oldDocMap[key] = doc
	}

	// 创建新文档的映射
	newDocMap := make(map[string]types.APIDoc)
	for _, doc := range newDocs {
		key := g.getDocKey(doc)
		newDocMap[key] = doc
	}

	// 查找新增和变更的文档
	for key, newDoc := range newDocMap {
		if oldDoc, exists := oldDocMap[key]; exists {
			// 检查是否有变更
			if !g.docsEqual(oldDoc, newDoc) {
				diff.Changed = append(diff.Changed, DocumentChange{
					Old: oldDoc,
					New: newDoc,
				})
			}
		} else {
			// 新增的文档
			diff.Added = append(diff.Added, newDoc)
		}
	}

	// 查找删除的文档
	for key, oldDoc := range oldDocMap {
		if _, exists := newDocMap[key]; !exists {
			diff.Removed = append(diff.Removed, oldDoc)
		}
	}

	return diff
}

// getDocKey 获取文档的唯一标识
func (g *Generator) getDocKey(doc types.APIDoc) string {
	return fmt.Sprintf("%s:%s", doc.Method, g.getRouter(doc))
}

// getRouter 获取路由信息
func (g *Generator) getRouter(doc types.APIDoc) string {
	if doc.Router != "" {
		return doc.Router
	}
	if doc.URL != "" {
		return doc.URL
	}
	return ""
}

// docsEqual 比较两个文档是否相等
func (g *Generator) docsEqual(doc1, doc2 types.APIDoc) bool {
	// 比较关键字段
	if doc1.Title != doc2.Title ||
		doc1.Description != doc2.Description ||
		doc1.Method != doc2.Method ||
		g.getRouter(doc1) != g.getRouter(doc2) ||
		doc1.Catalog != doc2.Catalog ||
		doc1.Remark != doc2.Remark {
		return false
	}

	// 比较参数数量
	if len(doc1.Header) != len(doc2.Header) ||
		len(doc1.Query) != len(doc2.Query) ||
		len(doc1.FormData) != len(doc2.FormData) ||
		len(doc1.Body) != len(doc2.Body) ||
		len(doc1.ResponseHeader) != len(doc2.ResponseHeader) ||
		len(doc1.ResponseBody) != len(doc2.ResponseBody) {
		return false
	}

	// 比较参数内容（简化比较，实际可能需要更详细的比较）
	return g.paramsEqual(doc1.Header, doc2.Header) &&
		g.paramsEqual(doc1.Query, doc2.Query) &&
		g.paramsEqual(doc1.FormData, doc2.FormData) &&
		g.paramsEqual(doc1.Body, doc2.Body) &&
		g.responseParamsEqual(doc1.ResponseHeader, doc2.ResponseHeader) &&
		g.responseParamsEqual(doc1.ResponseBody, doc2.ResponseBody)
}

// paramsEqual 比较请求参数
func (g *Generator) paramsEqual(params1, params2 []types.RequestParam) bool {
	if len(params1) != len(params2) {
		return false
	}

	for i := range params1 {
		if params1[i].Name != params2[i].Name ||
			params1[i].Type != params2[i].Type ||
			params1[i].Require != params2[i].Require ||
			params1[i].Remark != params2[i].Remark {
			return false
		}
	}

	return true
}

// responseParamsEqual 比较响应参数
func (g *Generator) responseParamsEqual(params1, params2 []types.ResponseParam) bool {
	if len(params1) != len(params2) {
		return false
	}

	for i := range params1 {
		if params1[i].Name != params2[i].Name ||
			params1[i].Type != params2[i].Type ||
			params1[i].Remark != params2[i].Remark {
			return false
		}
	}

	return true
}

// DocumentDiff 文档差异
type DocumentDiff struct {
	Added   []types.APIDoc   `json:"added"`
	Removed []types.APIDoc   `json:"removed"`
	Changed []DocumentChange `json:"changed"`
}

// DocumentChange 文档变更
type DocumentChange struct {
	Old types.APIDoc `json:"old"`
	New types.APIDoc `json:"new"`
}

// HasChanges 检查是否有变更
func (diff *DocumentDiff) HasChanges() bool {
	return len(diff.Added) > 0 || len(diff.Removed) > 0 || len(diff.Changed) > 0
}

// GetSummary 获取变更摘要
func (diff *DocumentDiff) GetSummary() string {
	return fmt.Sprintf("新增: %d, 删除: %d, 修改: %d",
		len(diff.Added), len(diff.Removed), len(diff.Changed))
}
