package user

import (
	"encoding/json"
	"example/internal/server/model/user"
	"net/http"

	"example/internal/pkg/response"
)

// Login
// runapi
// @catalog 测试文档/用户相关
// @title 用户登录
// @description 用户登录的接口
// @method post
// @url {{host}}/api/login
// @body user.LoginRequest
// @response_body response.Response{data=user.LoginResponse}
// @remark 登录接口
func Login(w http.ResponseWriter, r *http.Request) {
	var request user.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		_ = response.Error(http.StatusBadRequest, err.Error()).Write(w)
		return
	}
	user, err := user.Login(&request)
	if err != nil {
		response.Error(http.StatusUnauthorized, err.Error()).Write(w)
		return
	}
	_ = response.Success(user).Write(w)
}
