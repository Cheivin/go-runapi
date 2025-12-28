package user

import "fmt"

type User struct {
	Id       int    `json:"id"`       // 用户ID
	Username string `json:"username"` // 用户名
}
type LoginRequest struct {
	Username string `json:"username"`       // 用户名
	Password string `json:"password"`       // 密码
	TOTP     string `json:"totp,omitempty"` // 双因子认证码（可选）
}
type LoginResponse struct {
	User  User   `json:"user"`  // 用户信息
	Token string `json:"token"` // 登录凭证
}

func Login(request *LoginRequest) (*LoginResponse, error) {
	if request.Username == "admin" && request.Password == "123456" {
		return &LoginResponse{
			User: User{
				Id:       1,
				Username: "admin",
			},
			Token: "123456",
		}, nil
	}
	return nil, fmt.Errorf("用户名或密码错误")
}
