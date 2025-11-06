package main

import (
	"log"

	"github.com/as-tanais/observy/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
