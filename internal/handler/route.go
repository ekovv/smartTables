package handler

import "github.com/gin-gonic/gin"

func Route(c *gin.Engine, h *Handler) {
	c.GET("/smartTable", h.GetHome)
	c.POST("/smartTable", h.PostHome)
	c.GET("/result", h.GetResult)
	c.GET("/", h.ConnectionGet)
	c.POST("/", h.ConnectionPost)
}
