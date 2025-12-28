package types

import (
	"encoding/json"
	"strings"
)

// APIDocToPageContent 将APIDoc转换为PageContent结构
func APIDocToPageContent(apiDoc APIDoc) PageContent {
	// 确定URL，优先使用URL，如果没有则使用Router
	url := apiDoc.URL
	if url == "" {
		url = apiDoc.Router
	}

	// 转换请求参数
	headers := convertRequestParams(apiDoc.Header)
	query := convertRequestParams(apiDoc.Query)

	// 确定请求参数模式
	var params Params
	if len(apiDoc.FormData) > 0 {
		params.Mode = "formdata"
		params.FormData = convertRequestParams(apiDoc.FormData)
	} else if len(apiDoc.Body) > 0 {
		params.Mode = "json"
		params.JSONDesc = convertRequestParams(apiDoc.Body)
	} else {
		params.Mode = "formdata"
		params.FormData = []Param{}
	}

	// 默认添加空的urlencoded和jsonDesc
	if params.URLEncoded == nil {
		params.URLEncoded = []Param{}
	}
	if params.JSONDesc == nil {
		params.JSONDesc = []Param{}
	}

	// 转换响应参数
	responseParamsDesc := convertResponseParams(apiDoc.ResponseBody)

	// 如果有响应头参数，也添加到响应参数中，并在remark中标注
	if len(apiDoc.ResponseHeader) > 0 {
		headerParams := convertResponseParamsWithRemark(apiDoc.ResponseHeader, "header参数")
		responseParamsDesc = append(responseParamsDesc, headerParams...)
	}

	return PageContent{
		PageTitle: apiDoc.Title,
		Info: Info{
			Title:       apiDoc.Title,
			Description: apiDoc.Description,
			Method:      apiDoc.Method,
			URL:         url,
		},
		Request: Request{
			Params:  params,
			Headers: headers,
			Query:   query,
		},
		Response: Response{
			ResponseParamsDesc: responseParamsDesc,
			Remark:             apiDoc.Remark,
		},
	}
}

// convertRequestParams 转换请求参数
func convertRequestParams(params []RequestParam) []Param {
	result := make([]Param, 0)
	for _, p := range params {
		param := Param{
			//Disable: "0",
			Name:    p.Name,
			Value:   "",
			Type:    p.Type,
			Require: p.Require,
			Remark:  p.Remark,
		}
		if param.Require == "true" {
			param.Require = "1"
		} else if param.Require == "false" {
			param.Require = "0"
		}
		result = append(result, param)
	}
	return result
}

// convertResponseParams 转换响应参数
func convertResponseParams(params []ResponseParam) []ResponseParamDesc {
	var result []ResponseParamDesc
	for _, p := range params {
		result = append(result, ResponseParamDesc{
			Name:   p.Name,
			Type:   p.Type,
			Remark: p.Remark,
		})
	}
	return result
}

// convertResponseParamsWithRemark 转换响应参数并添加备注
func convertResponseParamsWithRemark(params []ResponseParam, remarkPrefix string) []ResponseParamDesc {
	var result []ResponseParamDesc
	for _, p := range params {
		remark := p.Remark
		if remark != "" && !strings.Contains(remark, remarkPrefix) {
			remark = remarkPrefix + " - " + remark
		} else if remark == "" {
			remark = remarkPrefix
		}

		result = append(result, ResponseParamDesc{
			Name:   p.Name,
			Type:   p.Type,
			Remark: remark,
		})
	}
	return result
}

// MergeWithFullContent 将PageContent与PageContentFull合并
func MergeWithFullContent(base PageContent, full PageContentFull) PageContentFull {
	// 从base结构更新full结构
	full.PageTitle = base.PageTitle
	full.Info.Title = base.Info.Title
	full.Info.Description = base.Info.Description
	full.Info.Method = base.Info.Method
	full.Info.URL = base.Info.URL

	// 更新请求结构
	full.Request.Params.Mode = base.Request.Params.Mode
	full.Request.Params.URLEncoded = base.Request.Params.URLEncoded
	full.Request.Params.FormData = base.Request.Params.FormData
	full.Request.Params.JSONDesc = base.Request.Params.JSONDesc
	full.Request.Headers = base.Request.Headers
	full.Request.Query = base.Request.Query

	// 更新响应结构
	full.Response.ResponseParamsDesc = base.Response.ResponseParamsDesc
	full.Response.Remark = base.Response.Remark

	return full
}

// CreateDefaultFullContent 创建默认的完整内容结构
func CreateDefaultFullContent() PageContentFull {
	return PageContentFull{
		Info: InfoFull{
			From:      "runapi",
			Type:      "api",
			APIStatus: "0",
		},
		Request: RequestFull{
			Params: ParamsFull{
				Mode: "formdata",
			},
			Cookies: []Cookie{},
			Auth: Auth{
				Type:     "none",
				Disabled: "0",
			},
			PathVariable: []json.RawMessage{},
		},
		Response: ResponseFull{
			ResponseStatus: 200,
			ResponseTime:   0,
			ResponseSize:   0,
		},
		Scripts: Scripts{
			Pre:  "",
			Post: "",
		},
		TestCases: []json.RawMessage{},
		Extend:    json.RawMessage("{}"),
	}
}
