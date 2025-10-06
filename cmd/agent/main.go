package main

import (
	"time"

	"github.com/as-tanais/observy/internal/agent"
)

const pollInterval = time.Second * 2
const reportInterval = time.Second * 10

var gaugeNames = []string{"Alloc", "BuckHashSys"}
var counterNames = []string{"Alloc", "BuckHashSys"}

func main() {

	m := agent.Collect()

	agent.Send(m)
}
