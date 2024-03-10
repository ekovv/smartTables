package main

import (
	"smartTables/config"
	"smartTables/internal/handler"
	"smartTables/internal/service"
	"smartTables/internal/storage"
)

func main() {
	conf := config.New()
	stM, err := storage.NewPostgresDBStorage(conf)
	if err != nil {
		return
	}
	sr := service.NewService(stM, conf)
	h := handler.NewHandler(sr, conf)
	h.Start()

}
