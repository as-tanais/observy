package main

import (
	"log"
	_ "net/http/pprof"

	"github.com/as-tanais/observy/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
