package api

import (
	"signalingServer/pkg/app"

	"github.com/gin-gonic/gin"
)

type handler struct {
	app *app.Service
}

func NewHandler(app *app.Service) *handler {
	return &handler{
		app: app,
	}
}

func (h *handler) RegisterRoutes(g *gin.Engine) {
	g.GET("/ws/signaling", h.app.WebsocketSignalHandler)
	g.GET("/clients", h.app.ListSignalClients)
}
