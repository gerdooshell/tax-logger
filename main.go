package main

import (
	"flag"
	"fmt"
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
	logger := serviceLogger.GetServiceLoggerInstance()
	for i := 0; i < 100; i++ {
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

		err := logger.Log(log)
		fmt.Println("err:", err)
	}
	<-time.After(time.Second * 5)
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

func readEach() {
	serviceLogQueue := queue.NewQueue[entities.ServiceLog](10, false)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		fmt.Println(i)
		if err := serviceLogQueue.Insert(entities.ServiceLog{Message: strconv.Itoa(i)}); err != nil {
			fmt.Println(err)
			break
		}
		outChan := serviceLogQueue.Read()
		wg.Add(1)
		go readOut(outChan, &wg)
	}
	wg.Wait()
}

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
