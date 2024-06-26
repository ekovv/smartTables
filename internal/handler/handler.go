package handler

import (
	"crypto/rand"
	"fmt"
	createv1 "github.com/ekovv/protosDB/gen/go/creator"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
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
	if res == nil {
		c.HTML(http.StatusOK, "smartTables.html", gin.H{
			"message": "Запрос успешно выполнен",
		})
		return
	}
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

	c.Redirect(http.StatusMovedPermanently, "/")
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
	login := session.Get("login").(string)

	m, err := s.service.GetLastDB(c.Request.Context(), login)
	if err != nil {
		HandlerErr(c, err)
		return
	}

	c.HTML(http.StatusOK, "connections.html", gin.H{
		"buttons": m,
	})
}

func (s *Handler) ConnectionPost(c *gin.Context) {
	session := sessions.Default(c)
	if session.Get("authenticated") != true {
		c.Redirect(http.StatusMovedPermanently, "/login")
		return
	}
	login := session.Get("login").(string)
	dbName := c.PostForm("dbName")
	db := c.PostForm("database")
	strings.ToLower(db)
	connectionString := c.PostForm("connectionString")
	button := c.PostForm("button")
	value := c.PostForm(button)
	if button != "" {
		dbName = button
		connectionString = value
		typeDB, err := s.service.GetConnectionFromBtn(c.Request.Context(), login, connectionString, dbName)
		if err != nil {
			HandlerErr(c, err)
			return
		}
		session.Set("database", typeDB)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/smartTable")
		return
	}
	if db == "sqlite" {
		file, err := c.FormFile("sqliteDbFile")
		if err != nil {
			HandlerErr(c, err)
			return
		}

		s.service.GetConnectionWithFile(login, db, dbName, file)
		session.Set("database", db)
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/smartTable")
		return
	}
	session.Set("database", db)
	session.Save()
	s.service.GetConnection(c.Request.Context(), login, db, connectionString, dbName)

	c.Redirect(http.StatusMovedPermanently, "/smartTable")
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
	err := s.service.Logout(login)
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

func (s *Handler) CreateDatabase(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("login").(string)
	login := c.PostForm("login")
	password := c.PostForm("password")
	dbName := c.PostForm("databaseName")
	dbType := c.PostForm("databaseForGRPC")

	conn, err := grpc.DialContext(c.Request.Context(), s.config.HostGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()

	// Создайте новый клиент gRPC.
	client := createv1.NewCreatorClient(conn)

	// Вызовите метод CreateDB.
	response, err := client.CreateDB(c.Request.Context(), &createv1.CreateDBRequest{
		User:     user,
		Login:    login,
		Password: password,
		DbName:   dbName,
		DbType:   dbType,
	})
	if err != nil {
		log.Fatalf("Ошибка при вызове CreateDB: %v", err)
	}
	c.HTML(http.StatusOK, "connections.html", gin.H{
		"data": response.GetConnectionString(),
	})

}
