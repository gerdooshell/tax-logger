package main

import (
	"flag"
	"fmt"
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
	go report()
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

func report() {
	t := time.NewTicker(time.Second * 5)
	var memStats runtime.MemStats

	for {
		runtime.ReadMemStats(&memStats)
		fmt.Println(runtime.NumGoroutine(), memStats.Alloc/1024, memStats.HeapAlloc/1024, memStats.NumGC)
		<-t.C
	}
}

func runGC() {
	ticker := time.NewTicker(time.Minute)
	for {
		<-ticker.C
		runtime.GC()
	}
}
