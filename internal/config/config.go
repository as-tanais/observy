package config

import (
	"log"
	"os"
)

type ServerConfig struct {
	ServerAddress string
}

func Load(addrFlag string) (*ServerConfig, error) {

	log.Println("FLAG приехал в LOAD", addrFlag)

	addr := os.Getenv("ADDRESS")
	if addr == "" {
		log.Println("Переменная ADDRESS не указана")
		addr = addrFlag
	}
	if addr == "" {
		log.Println("Флаг не задан")
		addr = "localhost:8080"
	}
	return &ServerConfig{ServerAddress: addr}, nil
}
