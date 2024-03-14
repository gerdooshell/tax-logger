package main

import (
	"flag"
	"runtime"
	"time"

	"github.com/gerdooshell/tax-logger/controller/server"
	"github.com/gerdooshell/tax-logger/environment"
)

func main() {
	env := readEnvironment()
	if err := environment.SetEnvironment(env); err != nil {
		panic(err)
	}
	go runGC()
	err := server.ServeGRPC()
	panic(err)
}

func readEnvironment() environment.Environment {
	isProdEnvPtr := flag.Bool("prod", false, "is environment prod")
	flag.Parse()
	env := environment.Dev
	if *isProdEnvPtr {
		env = environment.Prod
	}
	return env
}

func runGC() {
	ticker := time.NewTicker(time.Minute)
	for {
		<-ticker.C
		runtime.GC()
	}
}
