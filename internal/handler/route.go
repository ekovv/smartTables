package handler

import "github.com/gin-gonic/gin"

func Route(c *gin.Engine, h *Handler) {
	c.GET("/smartTable", h.GetHome)
	c.POST("/result", h.GetHome)
}
