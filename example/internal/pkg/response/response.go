package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Code int    `json:"code"` // 状态码
	Msg  string `json:"msg"`  // 提示信息
	Data any    `json:"data"` //  数据
}

func (r *Response) Write(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(r)
}

func Error(code int, msg string) *Response {
	return &Response{
		Code: code,
		Msg:  msg,
	}
}

func Success(data any) *Response {
	return &Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	}
}
