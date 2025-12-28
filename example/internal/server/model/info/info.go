package info

import "example/internal/server/model/user"

type (
	InfoRequest struct {
	}
	InfoResponse struct {
		User User `json:"user"`
	}
	User struct {
		user.User `json:"user"`
		Avatar    string `json:"avatar"` // 头像
	}
)

func GetInfo(request *InfoRequest) (*InfoResponse, error) {
	return &InfoResponse{
		User: User{
			User: user.User{
				Id:       1,
				Username: "admin",
			},
			Avatar: "https://example.com/avatar.png",
		},
	}, nil
}
