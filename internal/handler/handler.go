package handler

import (
	"crypto/rand"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
	"smartTables/config"
	"smartTables/internal/domains"
	"strings"
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

	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	store := cookie.NewStore(key)
	router.Use(sessions.Sessions("token", store))

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
	session := sessions.Default(c)
	if session.Get("authenticated") != true {
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}

	c.HTML(http.StatusOK, "smartTables.html", nil)
}

func (s *Handler) PostHome(c *gin.Context) {
	ctx := c.Request.Context()
	session := sessions.Default(c)
	if session.Get("authenticated") != true {
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}
	query := c.PostForm("query")
	login := session.Get("login").(string)
	res, err := s.service.ExecQuery(ctx, query, login)
	if err != nil {
		HandlerErr(c, err)
		return
	}
	err = s.service.SaveQuery(ctx, query, login)
	if err != nil {
		HandlerErr(c, err)
		return
	}

	c.HTML(http.StatusOK, "result.html", gin.H{
		"data": res,
	})
}

func (s *Handler) GetResult(c *gin.Context) {
	c.HTML(http.StatusOK, "result.html", nil)

}

func (s *Handler) LoginPost(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("authenticated") == true {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	login := c.PostForm("login")
	err := s.service.Login(c.Request.Context(), login, c.PostForm("password"))
	if err != nil {
		HandlerErr(c, err)
		return
	}

	session.Set("authenticated", true)
	session.Set("login", login)
	session.Options(sessions.Options{MaxAge: 60 * 60})
	session.Save()

	c.HTML(http.StatusOK, "connections.html", nil)
}

func (s *Handler) Login(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("authenticated") == true {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}
	c.HTML(http.StatusOK, "login.html", nil)
}

func (s *Handler) RegistrationPost(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("authenticated") == true {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	ctx := c.Request.Context()
	err := s.service.Registration(ctx, c.PostForm("login"), c.PostForm("password"))
	if err != nil {
		HandlerErr(c, err)
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/login")

}

func (s *Handler) Registration(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("authenticated") == true {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	c.HTML(http.StatusOK, "registration.html", nil)
}

func (s *Handler) ConnectionGet(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("authenticated") != true {
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}

	c.HTML(http.StatusOK, "connections.html", nil)
}

func (s *Handler) ConnectionPost(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("authenticated") != true {
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}

	dbName := c.PostForm("dbName")
	db := c.PostForm("database")
	strings.ToLower(db)
	login := session.Get("login").(string)
	if db == "sqlite" {
		file, err := c.FormFile("sqliteDbFile")
		if err != nil {
			HandlerErr(c, err)
			return
		}

		s.service.GetConnectionWithFile(login, db, dbName, file)
		session.Set("database", db)
		session.Save()
		c.HTML(http.StatusOK, "smartTables.html", nil)
		return
	}

	s.service.GetConnection(login, db, c.PostForm("connectionString"), dbName)

	c.HTML(http.StatusOK, "smartTables.html", nil)
}

func (s *Handler) ShowTables(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("authenticated") != true {
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}
	login := session.Get("login").(string)
	data, err := s.service.GetTables(c.Request.Context(), login)
	if err != nil {
		HandlerErr(c, err)
		return
	}

	c.HTML(http.StatusOK, "allTables.html", gin.H{
		"data": data,
	})
}

func (s *Handler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	login := session.Get("login").(string)
	db := session.Get("database")
	dbValue := ""
	if db == nil {
		err := s.service.Logout(login, dbValue)
		if err != nil {
			HandlerErr(c, err)
			return
		}
		session.Clear()
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/login")
	} else {
		dbValue = db.(string)
	}
	err := s.service.Logout(login, dbValue)
	if err != nil {
		HandlerErr(c, err)
		return
	}
	session.Clear()
	session.Save()
	c.Redirect(http.StatusMovedPermanently, "/login")
}

func (s *Handler) GetFile(c *gin.Context) {
	session := sessions.Default(c)
	file, err := c.FormFile("fileUpload")
	if err != nil {
		HandlerErr(c, err)
	}
	login := session.Get("login").(string)
	res, err := s.service.QueryFromFile(c.Request.Context(), file, login)
	if err != nil {
		HandlerErr(c, err)
		return
	}

	c.HTML(http.StatusOK, "result.html", gin.H{
		"data": res,
	})
}

func (s *Handler) GetHistory(c *gin.Context) {
	session := sessions.Default(c)
	login := session.Get("login").(string)
	res, err := s.service.GetHistory(c.Request.Context(), login)
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

	c.HTML(http.StatusOK, "history.html", gin.H{
		"data": data,
	})
}

func (s *Handler) SwitchDatabase(c *gin.Context) {
	session := sessions.Default(c)
	login := session.Get("login").(string)
	db := session.Get("database").(string)
	err := s.service.Switch(login, db)
	if err != nil {
		HandlerErr(c, err)
		return
	}
	session.Delete("database")
	session.Save()
	c.Redirect(http.StatusMovedPermanently, "/")

}
