package config

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
)

type Config struct {
	Host     string `json:"host"`
	HostGRPC string `json:"hostGRPC"`
	DB       string `json:"dsn"`
	Salt     string `json:""`
	CFile    string
}

type F struct {
	host     *string
	hostGRPC *string
	db       *string
	salt     *string
	cFile    *string
}

var f F

const addr = ":8080"
const addrGRPC = ":44044"

func init() {
	f.host = flag.String("a", addr, "-a=")
	f.hostGRPC = flag.String("g", addrGRPC, "-g=")
	f.db = flag.String("d", "", "-d=db")
	f.salt = flag.String("s", "", "-s=salt")
	f.cFile = flag.String("c", "", "-c=")

}

func New() (c Config) {
	flag.Parse()
	if envHost := os.Getenv("HOST"); envHost != "" {
		f.host = &envHost
	}
	if envHostGRPC := os.Getenv("HOSTGRPC"); envHostGRPC != "" {
		f.hostGRPC = &envHostGRPC
	}
	if envDB := os.Getenv("DB_CONNECTION_STRING"); envDB != "" {
		f.db = &envDB
	}
	if envSalt := os.Getenv("SALT"); envSalt != "" {
		f.salt = &envSalt
	}
	c.Host = *f.host
	c.HostGRPC = *f.hostGRPC
	c.DB = *f.db
	c.Salt = *f.salt
	c.CFile = *f.cFile
	file, err := os.Open(c.CFile)
	if err != nil {
		log.Fatalf("Не удалось открыть файл: %v", err)
		return
	}
	defer file.Close()

	all, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Ошибка при чтении файла: %v", err)
		return
	}

	err = json.Unmarshal(all, &c)
	if err != nil {
		log.Fatalf("Ошибка при разборе JSON: %v", err)
		return
	}
	return c
}
