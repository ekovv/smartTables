package handler

import (
	"fmt"
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
		HandlerErr(c, err)
		return
	}
	data := make([][]string, len(res))
	for i, row := range res {
		data[i] = make([]string, len(row))
		for j, col := range row {
			data[i][j] = fmt.Sprint(col)
		}
	}

	c.HTML(http.StatusOK, "result.html", gin.H{
		"data": data,
	})
}

func (s *Handler) GetResult(c *gin.Context) {
	c.HTML(http.StatusOK, "result.html", nil)

}

func (s *Handler) LoginPost(c *gin.Context) {

}

func (s *Handler) Login(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func (s *Handler) RegistrationPost(c *gin.Context) {
	ctx := c.Request.Context()
	err := s.service.Registration(ctx, c.PostForm("login"), c.PostForm("password"))
	if err != nil {
		HandlerErr(c, err)
		return
	}
}

func (s *Handler) Registration(c *gin.Context) {
	c.HTML(http.StatusOK, "registration.html", nil)
}
