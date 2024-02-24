package main

import (
	"flag"
	"fmt"
	"github.com/gerdooshell/tax-logger/controller/server"
	"github.com/gerdooshell/tax-logger/entities"
	"github.com/gerdooshell/tax-logger/entities/severity"
	"github.com/gerdooshell/tax-logger/environment"
	serviceLogger "github.com/gerdooshell/tax-logger/interactors/service-logger"
	"github.com/gerdooshell/tax-logger/lib/queue"
	"runtime"
	"strconv"
	"sync"
	"time"
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

func run() {
	logger := serviceLogger.GetServiceLoggerInstance()
	//ticker := time.NewTicker(time.Microsecond * 1000)
	t0 := time.Now()
	const count = 100_000_000
	for i := 0; i < count; i++ {
		//<-time.After(time.Microsecond * 1)
		//<-ticker.C
		log := entities.ServiceLog{
			Message:   "hello logger",
			Severity:  severity.Info,
			Timestamp: time.Now(),
			Origin: entities.OriginLog{
				ProcessId:    strconv.Itoa(i),
				ServiceName:  "manual",
				FunctionName: "main",
				StackTrace:   "custom stack trace",
			},
		}

		if err := logger.Log(log); err != nil {
			fmt.Println("err:", err)
		}
	}
	fmt.Println("inserted in:", time.Since(t0))
	<-time.After(time.Second * 560)
	fmt.Println("processed in:", time.Since(t0))
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

//func readEach() {
//	serviceLogQueue := queue.NewQueue[entities.ServiceLog](10, false)
//	wg := sync.WaitGroup{}
//	for i := 0; i < 10; i++ {
//		fmt.Println(i)
//		if err := serviceLogQueue.Insert(entities.ServiceLog{Message: strconv.Itoa(i)}); err != nil {
//			fmt.Println(err)
//			break
//		}
//		outChan := serviceLogQueue.Read()
//		wg.Add(1)
//		go readOut(outChan, &wg)
//	}
//	wg.Wait()
//}

func readOut(outChan <-chan queue.Output[entities.ServiceLog], wg *sync.WaitGroup) {
	out := <-outChan
	value := out.Value
	err := out.Err
	fmt.Println(value, err)
	<-time.After(time.Second * 5)
	out.IsDone(true)
	fmt.Println(value.Message, "Done", runtime.NumGoroutine())
	wg.Done()
}
