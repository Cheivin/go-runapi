package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"example/internal/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	go func() {
		fmt.Println("server start")
		_ = http.ListenAndServe(":8080", server.GetEngine())
	}()
	<-ctx.Done()
	fmt.Println("server exit")
}
