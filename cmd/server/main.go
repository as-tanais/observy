package main

import (
	"log"
	_ "net/http/pprof"

	"github.com/as-tanais/observy/internal/buildinfo"
	"github.com/as-tanais/observy/internal/server"
)

var buildVersion string
var buildDate string
var buildCommit string

func main() {

	buildinfo.PrintInfo(buildVersion, buildDate, buildCommit)

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
