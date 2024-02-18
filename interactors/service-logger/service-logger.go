package serviceLogger

import (
	"fmt"
	postgresService "github.com/gerdooshell/tax-logger/data-access/postgres_service"
	"github.com/gerdooshell/tax-logger/entities"
	"github.com/gerdooshell/tax-logger/interactors"
	"github.com/gerdooshell/tax-logger/lib/queue"
	"sync"
	"time"
)

type ServiceLogger interface {
	Log(serviceLog entities.ServiceLog) error
}

var serviceLoggerInstance *serviceLoggerImpl

func GetServiceLoggerInstance() ServiceLogger {
	if serviceLoggerInstance != nil {
		return serviceLoggerInstance
	}
	bufferSize := 10000
	serviceLoggerInstance = &serviceLoggerImpl{
		logQueue:                 queue.NewQueue[entities.ServiceLog](bufferSize, true),
		dataService:              postgresService.NewPostgresService(),
		bulkInsertCountThreshold: 100,
		bulkInsertTimeThreshold:  time.Second * 10,
	}
	go serviceLoggerInstance.persist()
	return serviceLoggerInstance
}

type serviceLoggerImpl struct {
	logQueue                 queue.Queue[entities.ServiceLog]
	dataService              interactors.DataAccess
	countQueued              int
	bulkInsertCountThreshold int
	bulkInsertTimeThreshold  time.Duration
	firstElementInsertTime   time.Time
}

var muLogBulk sync.Mutex

func (s *serviceLoggerImpl) LogBulk(serviceLogs []entities.ServiceLog) (err error) {
	for _, log := range serviceLogs {
		if err = log.Validate(); err != nil {
			return
		}
	}
	muLogBulk.Lock()
	defer muLogBulk.Unlock()
	for _, log := range serviceLogs {
		if err = s.logQueue.Insert(log); err != nil {
			return
		}
		if s.countQueued == 0 {
			s.firstElementInsertTime = time.Now()
		}
		s.countQueued++
	}
	return
}

var muLog sync.Mutex

func (s *serviceLoggerImpl) Log(serviceLog entities.ServiceLog) (err error) {
	if err = serviceLog.Validate(); err != nil {
		return
	}
	muLog.Lock()
	defer muLog.Unlock()
	if err = s.logQueue.Insert(serviceLog); err != nil {
		return
	}
	if s.countQueued == 0 {
		s.firstElementInsertTime = time.Now()
	}
	s.countQueued++
	return
}

func (s *serviceLoggerImpl) persist() {
	outChan := s.logQueue.ReadAll()
	for out := range outChan {
		fmt.Println(out.Value)
		if err := <-s.dataService.SaveServiceLogs([]entities.ServiceLog{out.Value}); err != nil {
			out.IsDone(false)
		}
		out.IsDone(true)
	}
}
