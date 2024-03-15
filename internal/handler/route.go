package handler

import "github.com/gin-gonic/gin"

func Route(c *gin.Engine, h *Handler) {
	c.GET("/smartTable", h.GetHome)
	c.POST("/smartTable", h.PostHome)
	c.GET("/result", h.GetResult)
	c.GET("/", h.ConnectionGet)
	c.POST("/", h.ConnectionPost)
	c.GET("/registration", h.Registration)
	c.POST("/registration", h.RegistrationPost)
	c.GET("/login", h.Login)
	c.POST("/login", h.LoginPost)
	c.GET("/tables", h.ShowTables)
	c.POST("/logout", h.Logout)
	c.POST("/upload", h.GetFile)
	c.GET("/history", h.GetHistory)
	c.POST("/switch", h.SwitchDatabase)
}
