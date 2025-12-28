package parser

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/cheivin/go-runapi/pkg/types"
)

// Parser 文档解析器
type Parser struct {
	fset           *token.FileSet
	structInfos    map[string]types.StructInfo  // key: "package.Struct"
	packageImports map[string]map[string]string // map[filePath]map[alias]packagePath
	packagePaths   map[string]string            // map[packageName]packagePath
	packageDir     string
	extraDirs      []string
	includeVendor  bool
}

// NewParser 创建新的解析器
func NewParser(docScanDir string, structScanDirs []string, includeVendor bool) *Parser {
	return &Parser{
		fset:           token.NewFileSet(),
		structInfos:    make(map[string]types.StructInfo),
		packageImports: make(map[string]map[string]string),
		packagePaths:   make(map[string]string),
		packageDir:     docScanDir,     // 文档扫描目录
		extraDirs:      structScanDirs, // 结构体扫描目录列表
		includeVendor:  includeVendor,
	}
}

// ParseDir 解析指定目录
func (p *Parser) ParseDir() ([]types.APIDoc, error) {
	var apiDocs []types.APIDoc

	// 首先解析所有结构体信息
	err := p.parseStructs()
	if err != nil {
		return nil, fmt.Errorf("解析结构体失败: %v", err)
	}

	// 然后解析API文档
	err = filepath.Walk(p.packageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		docs, err := p.parseFile(path)
		if err != nil {
			return fmt.Errorf("解析文件 %s 失败: %v", path, err)
		}

		apiDocs = append(apiDocs, docs...)
		return nil
	})

	return apiDocs, err
}

// parseStructs 解析所有结构体定义
func (p *Parser) parseStructs() error {
	// 要扫描的所有目录：文档目录 + 结构体目录
	dirsToScan := append([]string{p.packageDir}, p.extraDirs...)

	// 去重，避免重复扫描同一目录
	seen := make(map[string]bool)
	var uniqueDirs []string
	for _, dir := range dirsToScan {
		if !seen[dir] {
			seen[dir] = true
			uniqueDirs = append(uniqueDirs, dir)
		}
	}

	for _, dir := range uniqueDirs {
		err := p.parseStructsInDir(dir)
		if err != nil {
			return fmt.Errorf("解析目录 %s 中的结构体失败: %v", dir, err)
		}
	}

	return nil
}

// parseImports 解析文件的导入信息
func (p *Parser) parseImports(filePath string, file *ast.File) {
	imports := make(map[string]string)

	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		var alias string
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			// 从导入路径中提取包名
			parts := strings.Split(importPath, "/")
			alias = parts[len(parts)-1]
		}

		imports[alias] = importPath
	}

	p.packageImports[filePath] = imports
}

// parseStructsInDir 在指定目录中解析结构体定义
func (p *Parser) parseStructsInDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过vendor目录（除非明确包含）
		if !p.includeVendor && strings.Contains(path, "vendor/") {
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		file, err := parser.ParseFile(p.fset, path, src, parser.ParseComments)
		if err != nil {
			return err
		}

		// 解析包导入信息
		p.parseImports(path, file)

		// 获取包的相对路径作为包路径标识
		relPath, err := filepath.Rel(p.packageDir, filepath.Dir(path))
		if err != nil {
			relPath = filepath.Dir(path)
		}
		packageName := file.Name.Name

		ast.Inspect(file, func(n ast.Node) bool {
			genDecl, ok := n.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				return true
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				// 使用包名+结构体名作为key
				key := packageName + "." + typeSpec.Name.Name
				structInfo := types.StructInfo{
					Name:        typeSpec.Name.Name,
					Package:     packageName,
					PackagePath: relPath,
				}

				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 {
						// 嵌入字段（匿名字段）
						fieldType := p.getTypeString(field.Type)
						param := types.ResponseParam{
							Name:     fieldType,
							Type:     fieldType,
							Required: true, // 嵌入字段默认必传
							Remark:   "嵌入字段",
						}

						// 提取字段注释
						if field.Comment != nil {
							for _, comment := range field.Comment.List {
								param.Remark = strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
							}
						}

						structInfo.Fields = append(structInfo.Fields, param)
					} else {
						// 普通字段
						for _, name := range field.Names {
							param := types.ResponseParam{
								Name: name.Name,
								Type: p.getTypeString(field.Type),
							}

							// 提取JSON tag
							if field.Tag != nil {
								tag := strings.Trim(field.Tag.Value, "`")
								if jsonName, omitempty, ok := p.extractJSONTagInfo(tag); ok {
									if jsonName == "-" {
										continue // 跳过不序列化的字段
									}
									param.Name = jsonName
									param.Required = !omitempty // 有omitempty则为非必传，否则为必传
								}
							}

							// 提取字段注释
							if field.Comment != nil {
								for _, comment := range field.Comment.List {
									param.Remark = strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
								}
							}

							structInfo.Fields = append(structInfo.Fields, param)
						}
					}
				}

				p.structInfos[key] = structInfo
			}

			return true
		})

		return nil
	})
}

// parseFile 解析单个文件
func (p *Parser) parseFile(filePath string) ([]types.APIDoc, error) {
	var apiDocs []types.APIDoc

	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(p.fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if funcDecl.Doc == nil {
			return true
		}

		// 检查是否包含 runapi 标记
		hasRunAPI := false
		for _, comment := range funcDecl.Doc.List {
			if strings.TrimSpace(comment.Text) == "// runapi" {
				hasRunAPI = true
				break
			}
		}

		if !hasRunAPI {
			return true
		}

		apiDoc, err := p.parseFuncDoc(funcDecl.Doc, filePath)
		if err != nil {
			fmt.Printf("解析函数 %s 的文档失败: %v\n", funcDecl.Name.Name, err)
			return true
		}

		// 添加位置信息用于错误定位
		apiDoc.FilePath = filePath
		apiDoc.FunctionName = funcDecl.Name.Name

		apiDocs = append(apiDocs, *apiDoc)
		return true
	})

	return apiDocs, nil
}

// parseFuncDoc 解析函数文档注释
func (p *Parser) parseFuncDoc(doc *ast.CommentGroup, filePath string) (*types.APIDoc, error) {
	apiDoc := &types.APIDoc{}

	for _, comment := range doc.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

		if text == "runapi" {
			continue
		}

		if strings.HasPrefix(text, "@remark") {
			apiDoc.Remark = strings.TrimSpace(strings.TrimPrefix(text, "@remark"))
			continue
		}

		// 使用空格分割，但保留第一个词作为key
		spaceIndex := strings.Index(text, " ")
		if spaceIndex == -1 {
			continue
		}

		key := text[:spaceIndex]
		value := strings.TrimSpace(text[spaceIndex:])

		if value == "" {
			continue
		}

		switch key {
		case "@catalog":
			apiDoc.Catalog = value
		case "@title":
			apiDoc.Title = value
		case "@description":
			apiDoc.Description = value
		case "@method":
			apiDoc.Method = value
		case "@router":
			apiDoc.Router = value
		case "@url":
			apiDoc.URL = value
		case "@param":
			paramParts := strings.Fields(value)
			if len(paramParts) >= 4 {
				paramName := paramParts[0]
				paramLocation := paramParts[1]
				paramType := paramParts[2]
				paramRequired := paramParts[3]
				paramRemark := ""
				if len(paramParts) > 4 {
					paramRemark = strings.Join(paramParts[4:], " ")
				}

				// 应用类型映射
				mappedType := p.mapGoTypeToRequestType(paramType)

				param := types.RequestParam{
					Name:    paramName,
					Type:    mappedType,
					Require: paramRequired,
					Remark:  paramRemark,
				}

				switch paramLocation {
				case "header":
					apiDoc.Header = append(apiDoc.Header, param)
				case "query":
					apiDoc.Query = append(apiDoc.Query, param)
				case "formData":
					apiDoc.FormData = append(apiDoc.FormData, param)
				}
			}
		case "@response":
			responseParts := strings.Fields(value)
			if len(responseParts) >= 3 {
				paramName := responseParts[0]
				paramLocation := responseParts[1]
				paramType := responseParts[2]
				paramRemark := ""
				if len(responseParts) > 3 {
					paramRemark = strings.Join(responseParts[3:], " ")
				}

				// 应用类型映射
				mappedType := p.mapGoTypeToResponseType(paramType)

				param := types.ResponseParam{
					Name:   paramName,
					Type:   mappedType,
					Remark: paramRemark,
				}

				if paramLocation == "header" {
					apiDoc.ResponseHeader = append(apiDoc.ResponseHeader, param)
				} else if paramLocation == "body" {
					apiDoc.ResponseBody = append(apiDoc.ResponseBody, param)
				}
			} else {
				// 处理结构体格式的响应
				responseValue := value
				if strings.Contains(responseValue, "{") && strings.HasSuffix(responseValue, "}") {
					// 解析嵌套响应格式
					nestedParams, err := p.parseNestedResponse(responseValue, filePath)
					if err == nil && len(nestedParams) > 0 {
						apiDoc.ResponseBody = append(apiDoc.ResponseBody, nestedParams...)
					}
				} else if _, exists := p.structInfos[responseValue]; exists {
					nestedParams := p.deepParseStruct(responseValue, "")
					apiDoc.ResponseBody = append(apiDoc.ResponseBody, nestedParams...)
				}
			}
		case "@response_body":
			responseValue := value
			if strings.Contains(responseValue, "{") && strings.HasSuffix(responseValue, "}") {
				// 解析嵌套响应格式
				nestedParams, err := p.parseNestedResponse(responseValue, filePath)
				if err == nil && len(nestedParams) > 0 {
					apiDoc.ResponseBody = append(apiDoc.ResponseBody, nestedParams...)
				}
			} else {
				// 尝试解析带包名的结构体引用
				structKey, err := p.resolveStructReference(responseValue, filePath)
				if err != nil {
					// 如果解析失败，尝试直接查找
					structKey = responseValue
				}

				if _, exists := p.structInfos[structKey]; exists {
					nestedParams := p.deepParseStruct(structKey, "")
					apiDoc.ResponseBody = append(apiDoc.ResponseBody, nestedParams...)
				}
			}
		case "@body":
			bodyType := value
			// 尝试解析带包名的结构体引用
			structKey, err := p.resolveStructReference(bodyType, filePath)
			if err != nil {
				// 如果解析失败，尝试直接查找
				structKey = bodyType
			}

			if _, exists := p.structInfos[structKey]; exists {
				nestedParams := p.deepParseStruct(structKey, "")
				// 转换为types.RequestParam并应用请求类型映射
				for _, field := range nestedParams {
					var requireStr string
					if field.Required {
						requireStr = "true"
					} else {
						requireStr = "false"
					}

					requestParam := types.RequestParam{
						Name:    field.Name,
						Type:    p.mapGoTypeToRequestType(field.Type),
						Require: requireStr,
						Remark:  field.Remark,
					}
					apiDoc.Body = append(apiDoc.Body, requestParam)
				}
			} else {
				fmt.Printf("警告: 未找到结构体 %s (原始: %s)\n", structKey, bodyType)
			}
		}
	}

	return apiDoc, nil
}

// parseParam 解析参数行
func (p *Parser) parseParam(paramStr, remark string) (*types.RequestParam, error) {
	// param格式: name type required
	parts := strings.Fields(paramStr)
	if len(parts) < 3 {
		return nil, fmt.Errorf("参数格式错误: %s", paramStr)
	}

	// 应用类型映射
	mappedType := p.mapGoTypeToRequestType(parts[1])

	return &types.RequestParam{
		Name:    parts[0],
		Type:    mappedType,
		Require: parts[2],
		Remark:  remark,
	}, nil
}

// extractJSONTag 提取JSON tag
func (p *Parser) extractJSONTag(tagStr string) string {
	name, _, _ := p.extractJSONTagInfo(tagStr)
	return name
}

// JSONTagInfo JSON标签信息
type JSONTagInfo struct {
	Name      string // 字段名
	Omitempty bool   // 是否包含omitempty
}

// extractJSONTagInfo 提取JSON tag的完整信息
func (p *Parser) extractJSONTagInfo(tagStr string) (string, bool, bool) {
	parts := strings.Split(tagStr, " ")
	if len(parts) == 0 {
		return "", false, false
	}

	jsonPart := parts[0]
	if strings.HasPrefix(jsonPart, "json:") {
		jsonTag := strings.Trim(jsonPart[5:], "\"")
		if jsonTag == "-" {
			return "-", false, false // 跳过不序列化的字段
		}

		var fieldName string
		var omitempty bool

		// 解析逗号分隔的选项
		if commaIndex := strings.Index(jsonTag, ","); commaIndex != -1 {
			fieldName = jsonTag[:commaIndex]
			options := jsonTag[commaIndex+1:]
			omitempty = strings.Contains(options, "omitempty")
		} else {
			fieldName = jsonTag
			omitempty = false
		}

		return fieldName, omitempty, true
	}

	return "", false, false
}

// getTypeString 获取类型字符串
func (p *Parser) getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + p.getTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + p.getTypeString(t.Elt)
	case *ast.SelectorExpr:
		return p.getTypeString(t.X) + "." + t.Sel.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// resolveStructReference 解析结构体引用，支持包名
func (p *Parser) resolveStructReference(structRef string, filePath string) (string, error) {
	// 如果没有包名前缀，尝试在当前包中查找
	if !strings.Contains(structRef, ".") {
		// 获取当前文件的包名
		if file, err := parser.ParseFile(p.fset, filePath, nil, parser.PackageClauseOnly); err == nil {
			currentPackage := file.Name.Name
			currentKey := currentPackage + "." + structRef
			if _, exists := p.structInfos[currentKey]; exists {
				return currentKey, nil
			}
		}

		// 如果当前包中没有找到，返回原始名称（可能是旧格式的兼容）
		return structRef, nil
	}

	parts := strings.SplitN(structRef, ".", 2)
	packageAlias := parts[0]
	structName := parts[1]

	// 获取当前文件的导入信息
	imports, exists := p.packageImports[filePath]
	if !exists {
		return "", fmt.Errorf("无法找到文件 %s 的导入信息", filePath)
	}

	// 查找包别名对应的导入路径
	importPath, exists := imports[packageAlias]
	if !exists {
		return "", fmt.Errorf("无法找到包别名 %s 对应的导入路径", packageAlias)
	}

	// 尝试在已解析的结构体中查找匹配的结构体
	var candidates []string
	for key, structInfo := range p.structInfos {
		if structInfo.Name == structName {
			// 检查包路径是否匹配 - 支持多种匹配方式
			pathMatch := false

			// 1. 直接匹配
			if structInfo.PackagePath == importPath {
				pathMatch = true
			} else if strings.HasSuffix(importPath, structInfo.PackagePath) {
				// 2. 后缀匹配（导入路径可能包含模块前缀）
				pathMatch = true
			} else if strings.HasSuffix(structInfo.PackagePath, importPath) {
				// 3. 前缀匹配（相对路径匹配）
				pathMatch = true
			} else if p.removeModulePrefix(importPath) == structInfo.PackagePath {
				// 4. 去掉模块名后匹配
				pathMatch = true
			}

			if pathMatch {
				candidates = append(candidates, key)
				// 如果包名完全匹配，优先选择
				if structInfo.Package == packageAlias {
					return key, nil
				}
			}
		}
	}

	// 如果有多个候选，选择路径最匹配的
	if len(candidates) > 0 {
		if len(candidates) == 1 {
			return candidates[0], nil
		}

		// 选择路径最长的匹配（更具体的路径）
		bestMatch := candidates[0]
		for _, candidate := range candidates {
			if len(p.structInfos[candidate].PackagePath) > len(p.structInfos[bestMatch].PackagePath) {
				bestMatch = candidate
			}
		}
		return bestMatch, nil
	}

	return "", fmt.Errorf("无法找到结构体 %s.%s (导入路径: %s)", packageAlias, structName, importPath)
}

// removeModulePrefix 移除导入路径中的模块名前缀
func (p *Parser) removeModulePrefix(importPath string) string {
	// 如果导入路径包含 /，尝试移除第一部分（模块名）
	if strings.Contains(importPath, "/") {
		parts := strings.SplitN(importPath, "/", 2)
		if len(parts) > 1 {
			return parts[1]
		}
	}
	return importPath
}

// getAvailableStructs 获取所有可用的结构体，用于调试
func (p *Parser) getAvailableStructs() []string {
	var structs []string
	for key := range p.structInfos {
		structs = append(structs, key)
	}
	return structs
}

// parseNestedResponse 解析嵌套响应格式，如 Response{data=UserInfo} 或 Response{result=user.Info}
func (p *Parser) parseNestedResponse(responseValue string, filePath string) ([]types.ResponseParam, error) {
	// 检查是否是 StructName{...} 格式
	if !strings.Contains(responseValue, "{") || !strings.HasSuffix(responseValue, "}") {
		return nil, fmt.Errorf("不是有效的嵌套响应格式")
	}

	// 提取结构体名称和内部内容
	leftBrace := strings.Index(responseValue, "{")
	baseStructName := strings.TrimSpace(responseValue[:leftBrace])
	innerContent := strings.TrimPrefix(responseValue, baseStructName+"{")
	innerContent = strings.TrimSuffix(innerContent, "}")

	var params []types.ResponseParam

	// 解析基础结构体名称（可能包含包名）
	baseStructKey, err := p.resolveStructReference(baseStructName, filePath)
	if err != nil {
		// 如果解析失败，尝试直接查找
		baseStructKey = baseStructName
	}

	// 首先添加基础结构体的字段
	if structInfo, exists := p.structInfos[baseStructKey]; exists {
		for _, field := range structInfo.Fields {
			params = append(params, field)
		}
	}

	// 如果没有内部覆盖内容，直接返回基础结构体字段
	if innerContent == "" {
		return params, nil
	}

	// 解析字段覆盖，如 "data=UserInfo, result=user.Info"
	fields := strings.Split(innerContent, ",")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		// 解析 fieldName=StructName 格式
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			continue
		}

		fieldName := strings.TrimSpace(parts[0])
		structName := strings.TrimSpace(parts[1])

		// 解析结构体引用（可能包含包名）
		structKey, err := p.resolveStructReference(structName, filePath)
		if err != nil {
			// 如果解析失败，尝试直接查找
			structKey = structName
		}

		if _, exists := p.structInfos[structKey]; exists {
			// 查找并替换对应的字段
			found := false
			for i, param := range params {
				if param.Name == fieldName {
					// 替换为新的结构体字段
					params[i] = types.ResponseParam{Name: fieldName, Type: "object", Remark: param.Remark}

					// 深度添加结构体的子字段
					nestedParams := p.deepParseStruct(structKey, fieldName+".")
					params = append(params, nestedParams...)
					found = true
					break
				}
			}

			// 如果没有找到对应字段，作为新字段添加
			if !found {
				// 添加对象字段
				remark := fmt.Sprintf("%s信息", fieldName)
				params = append(params, types.ResponseParam{Name: fieldName, Type: "object", Remark: remark})

				// 深度添加结构体字段
				nestedParams := p.deepParseStruct(structKey, fieldName+".")
				params = append(params, nestedParams...)
			}
		}
	}

	return params, nil
}

// deepParseStruct 深度解析结构体字段，展开嵌套结构体
func (p *Parser) deepParseStruct(structName string, prefix string) []types.ResponseParam {
	var params []types.ResponseParam

	structInfo, exists := p.structInfos[structName]
	if !exists {
		return params
	}

	for _, field := range structInfo.Fields {
		// 检查是否是嵌入字段（字段名和类型相同）
		isEmbedded := (field.Name == field.Type)

		// 清理字段类型，移除指针符号
		cleanType := strings.TrimPrefix(field.Type, "*")

		// 检查字段类型是否是已知的结构体
		if _, isStruct := p.structInfos[cleanType]; isStruct {
			// 递归解析嵌套结构体
			fieldPrefix := prefix
			if !isEmbedded {
				fieldPrefix = prefix + field.Name + "."
			}
			nestedParams := p.deepParseStruct(cleanType, fieldPrefix)
			params = append(params, nestedParams...)
		} else {
			// 检查是否是带包名的结构体引用
			if strings.Contains(cleanType, ".") {
				// 尝试解析带包名的结构体
				found := false
				for key := range p.structInfos {
					if strings.HasSuffix(key, "."+cleanType) || key == cleanType {
						// 递归解析嵌套结构体
						fieldPrefix := prefix
						if !isEmbedded {
							fieldPrefix = prefix + field.Name + "."
						}
						nestedParams := p.deepParseStruct(key, fieldPrefix)
						params = append(params, nestedParams...)
						found = true
						break
					}
				}
				if !found {
					// 只有非嵌入字段才添加到参数列表
					if !isEmbedded {
						param := types.ResponseParam{
							Name:     prefix + field.Name,
							Type:     p.mapGoTypeToResponseType(field.Type),
							Required: field.Required,
							Remark:   field.Remark,
						}
						params = append(params, param)
					}
				}
			} else {
				// 尝试在同一包内查找结构体
				currentPackage := structInfo.Package
				prefixedKey := currentPackage + "." + cleanType
				if _, isStruct := p.structInfos[prefixedKey]; isStruct {
					// 递归解析嵌套结构体
					fieldPrefix := prefix
					if !isEmbedded {
						fieldPrefix = prefix + field.Name + "."
					}
					nestedParams := p.deepParseStruct(prefixedKey, fieldPrefix)
					params = append(params, nestedParams...)
				} else {
					// 只有非嵌入字段才添加到参数列表
					if !isEmbedded {
						param := types.ResponseParam{
							Name:     prefix + field.Name,
							Type:     p.mapGoTypeToResponseType(field.Type),
							Required: field.Required,
							Remark:   field.Remark,
						}
						params = append(params, param)
					}
				}
			}
		}
	}

	return params
}

// mapGoTypeToRequestType 将Go类型映射到请求参数类型
func (p *Parser) mapGoTypeToRequestType(goType string) string {
	// 处理指针类型
	if strings.HasPrefix(goType, "*") {
		goType = goType[1:]
	}

	// 处理数组类型
	if strings.HasPrefix(goType, "[]") {
		return "array"
	}

	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "uint", "uint8", "uint16", "uint32":
		return "int"
	case "int64", "uint64":
		return "long"
	case "float32":
		return "float"
	case "float64":
		return "double"
	case "bool":
		return "boolean"
	case "file":
		return "file"
	case "interface{}":
		return "object"
	default:
		// 对于自定义结构体，返回object
		return "object"
	}
}

// mapGoTypeToResponseType 将Go类型映射到响应参数类型
func (p *Parser) mapGoTypeToResponseType(goType string) string {
	// 处理指针类型
	if strings.HasPrefix(goType, "*") {
		goType = goType[1:]
	}

	// 处理数组类型
	if strings.HasPrefix(goType, "[]") {
		return "array"
	}

	switch goType {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "uint", "uint8", "uint16", "uint32":
		return "int"
	case "int64", "uint64":
		return "long"
	case "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "interface{}":
		return "object"
	default:
		// 对于自定义结构体，返回object
		return "object"
	}
}

// validateAPIDoc 校验API文档的必填字段
func (p *Parser) validateAPIDoc(doc types.APIDoc) error {
	location := fmt.Sprintf("文件: %s, 函数: %s", doc.FilePath, doc.FunctionName)

	if doc.Title == "" {
		return fmt.Errorf("%s - title字段是必填的\n修复建议: 在函数注释中添加 @title 接口标题", location)
	}
	if doc.Method == "" {
		return fmt.Errorf("%s - method字段是必填的\n修复建议: 在函数注释中添加 @method get|post|put|delete等HTTP方法", location)
	}
	if doc.Router == "" && doc.URL == "" {
		return fmt.Errorf("%s - router或url字段至少需要一个\n修复建议: 在函数注释中添加 @router /api/path 或 @url /api/path", location)
	}
	return nil
}

// GenerateJSON 生成JSON文档
func (p *Parser) GenerateJSON(apiDocs []types.APIDoc) (string, error) {
	var validationErrors []error

	// 校验所有API文档
	for _, doc := range apiDocs {
		if err := p.validateAPIDoc(doc); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	// 如果有校验错误，显示汇总信息
	if len(validationErrors) > 0 {
		fmt.Fprintf(os.Stderr, "发现 %d 个API文档校验问题:\n", len(validationErrors))
		for i, err := range validationErrors {
			fmt.Fprintf(os.Stderr, "%d. %v\n", i+1, err)
		}
		return "", fmt.Errorf("API文档校验失败，请修复上述问题后重试")
	}

	jsonData, err := json.MarshalIndent(apiDocs, "", "\t")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
