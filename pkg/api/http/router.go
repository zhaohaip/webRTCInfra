package http

import (
	"github.com/gin-gonic/gin"
)

type router struct {
	handler *Handler
}

func NewRouter(handler *Handler) *router {
	return &router{
		handler: handler,
	}
}

func (r *router) Run(addr string) error {
	g := gin.Default()
	r.registerRoutes(g)
	return g.Run(addr)
}

func (r *router) registerRoutes(g *gin.Engine) {
	g.GET("/ws/signaling", r.handler.WebsocketSignalHandler)
	g.GET("/clients", r.handler.ListSignalClients)
}
