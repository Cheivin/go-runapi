package server

import (
	"net/http"

	"example/internal/server/controller/info"
	"example/internal/server/controller/user"
)

func GetEngine() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", user.Login)
	mux.HandleFunc("/api/info", info.GetInfo)
	mux.HandleFunc("/api/avatar", info.SetAvatar)
	return mux
}
