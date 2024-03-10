package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"smartTables/config"
	"smartTables/internal/domains"
)

type Handler struct {
	service domains.Service
	engine  *gin.Engine
	config  config.Config
}

func NewHandler(service domains.Service, cnf config.Config) *Handler {
	router := gin.Default()
	router.LoadHTMLGlob("templates/html/*")
	h := &Handler{
		service: service,
		engine:  router,
		config:  cnf,
	}

	Route(router, h)
	return h
}

func (s *Handler) Start() {
	err := s.engine.Run(s.config.Host)
	if err != nil {
		return
	}
}

func (s *Handler) GetHome(c *gin.Context) {
	c.HTML(http.StatusOK, "smartTables.html", nil)
}

func (s *Handler) PostHome(c *gin.Context) {
	ctx := c.Request.Context()
	res, err := s.service.ExecQuery(ctx, c.PostForm("query"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.HTML(http.StatusOK, "result.html", gin.H{
		"data": data,
	})
}
