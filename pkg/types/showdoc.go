package types

import "encoding/json"

// PageContent ShowDoc页面内容结构（简化版）
type PageContent struct {
	PageTitle string   `json:"page_title"`
	Info      Info     `json:"info"`
	Request   Request  `json:"request"`
	Response  Response `json:"response"`
}

// Info 基本信息
type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Method      string `json:"method"`
	URL         string `json:"url"`
}

// Request 请求信息
type Request struct {
	Params  Params  `json:"params"`
	Headers []Param `json:"headers"`
	Query   []Param `json:"query"`
}

// Params 请求参数
type Params struct {
	Mode       string  `json:"mode"`
	URLEncoded []Param `json:"urlencoded"`
	FormData   []Param `json:"formdata"`
	JSONDesc   []Param `json:"jsonDesc"`
}

// Param 参数结构
type Param struct {
	//Disable string `json:"disable"`
	Name    string `json:"name"`
	Value   string `json:"value"`
	Type    string `json:"type"`
	Require string `json:"require"`
	Remark  string `json:"remark"`
}

// Response 响应信息
type Response struct {
	ResponseParamsDesc []ResponseParamDesc `json:"responseParamsDesc"`
	Remark             string              `json:"remark"`
}

// ResponseParamDesc 响应参数描述
type ResponseParamDesc struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Remark string `json:"remark"`
}

// PageContentFull ShowDoc页面内容完整结构
type PageContentFull struct {
	PageTitle string            `json:"page_title"`
	Info      InfoFull          `json:"info"`
	Request   RequestFull       `json:"request"`
	Response  ResponseFull      `json:"response"`
	Scripts   Scripts           `json:"scripts"`
	TestCases []json.RawMessage `json:"testCases"`
	Extend    json.RawMessage   `json:"extend"`
}

// InfoFull 完整信息
type InfoFull struct {
	From        string `json:"from"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Method      string `json:"method"`
	URL         string `json:"url"`
	Remark      string `json:"remark"`
	APIStatus   string `json:"apiStatus"`
}

// RequestFull 完整请求信息
type RequestFull struct {
	Params       ParamsFull        `json:"params"`
	Headers      []Param           `json:"headers"`
	Cookies      []Cookie          `json:"cookies"`
	Auth         Auth              `json:"auth"`
	Query        []Param           `json:"query"`
	PathVariable []json.RawMessage `json:"pathVariable"`
}

// ParamsFull 完整参数
type ParamsFull struct {
	Mode       string  `json:"mode"`
	URLEncoded []Param `json:"urlencoded"`
	FormData   []Param `json:"formdata"`
	JSON       string  `json:"json"`
	JSONDesc   []Param `json:"jsonDesc"`
}

// Cookie Cookie结构
type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Auth 认证信息
type Auth struct {
	Type     string `json:"type"`
	Disabled string `json:"disabled"`
}

// ResponseFull 完整响应信息
type ResponseFull struct {
	ResponseText           string              `json:"responseText"`
	ResponseOriginal       json.RawMessage     `json:"responseOriginal"`
	ResponseExample        string              `json:"responseExample"`
	ResponseHeader         json.RawMessage     `json:"responseHeader"`
	ResponseStatus         int                 `json:"responseStatus"`
	ResponseTime           int                 `json:"responseTime"`
	ResponseParamsDesc     []ResponseParamDesc `json:"responseParamsDesc"`
	ResponseFailExample    string              `json:"responseFailExample"`
	ResponseFailParamsDesc []ResponseParamDesc `json:"responseFailParamsDesc"`
	Remark                 string              `json:"remark"`
	ResponseSize           int                 `json:"responseSize"`
}

// Scripts 脚本信息
type Scripts struct {
	Pre  string `json:"pre"`
	Post string `json:"post"`
}
