package info

import (
	"encoding/json"
	"example/internal/pkg/response"
	"example/internal/server/model/info"
	"net/http"
)

// GetInfo
// runapi
// @catalog 测试文档/用户资料
// @title 获取用户信息
// @method get
// @router {{host}}/api/info
// @param token header string true 授权token
// @response_body response.Response{data=info.InfoResponse}
func GetInfo(w http.ResponseWriter, r *http.Request) {
	var request info.InfoRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		_ = response.Error(http.StatusBadRequest, err.Error()).Write(w)
		return
	}
	user, err := info.GetInfo(&request)
	if err != nil {
		response.Error(http.StatusUnauthorized, err.Error()).Write(w)
	}
	_ = response.Success(user).Write(w)
}

// SetAvatar
// runapi
// @catalog 测试文档/用户资料
// @title 设置头像
// @method post
// @router {{host}}/api/avatar
// @param token header string true 授权token
// @param avatar formData file true 头像
// @response_body response.Response
func SetAvatar(w http.ResponseWriter, r *http.Request) {
	_ = response.Success("").Write(w)
}
