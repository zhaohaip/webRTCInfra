package main

import (
	"signalingServer/pkg/api"
	"signalingServer/pkg/app"

	"github.com/gin-gonic/gin"
)

func main() {
	g := gin.Default()
	svc := app.NewService()

	h := api.NewHandler(svc)
	h.RegisterRoutes(g)

	g.Run(":8080")
}
