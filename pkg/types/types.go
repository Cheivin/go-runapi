package types

// RequestParam 表示请求参数的结构
type RequestParam struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Require string `json:"require"`
	Remark  string `json:"remark"`
}

// ResponseParam 表示响应参数的结构
type ResponseParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"` // 是否必传（基于omitempty标签）
	Remark   string `json:"remark"`
}

// APIDoc 表示一个API文档的结构
type APIDoc struct {
	Title          string          `json:"title"`
	Catalog        string          `json:"catalog"`
	Description    string          `json:"description"`
	Method         string          `json:"method"`
	Router         string          `json:"router,omitempty"`
	URL            string          `json:"url,omitempty"`
	Header         []RequestParam  `json:"header,omitempty"`
	Query          []RequestParam  `json:"query,omitempty"`
	FormData       []RequestParam  `json:"formData,omitempty"`
	Body           []RequestParam  `json:"body,omitempty"`
	ResponseHeader []ResponseParam `json:"response_header,omitempty"`
	ResponseBody   []ResponseParam `json:"response_body,omitempty"`
	Remark         string          `json:"remark,omitempty"`
	// 内部使用，不序列化到JSON
	FilePath     string `json:"-"`
	FunctionName string `json:"-"`
}

// StructInfo 表示结构体信息
type StructInfo struct {
	Name        string
	Package     string // 结构体所属的包名
	PackagePath string // 包的完整路径
	Fields      []ResponseParam
}
