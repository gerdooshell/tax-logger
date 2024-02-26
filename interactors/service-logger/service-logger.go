package serviceLogger

import (
	postgresService "github.com/gerdooshell/tax-logger/data-access/postgres_service"
	"github.com/gerdooshell/tax-logger/entities"
	"github.com/gerdooshell/tax-logger/interactors"
	queueBulk "github.com/gerdooshell/tax-logger/interactors/queue_bulk"
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
	const QueuePurgeTimeout = time.Second * 5
	serviceLoggerInstance = &serviceLoggerImpl{
		logQueue:    queueBulk.NewQueueBulk[entities.ServiceLog](3000, 10000, QueuePurgeTimeout),
		dataService: postgresService.NewPostgresService(),
	}
	go serviceLoggerInstance.persist()
	return serviceLoggerInstance
}

type serviceLoggerImpl struct {
	logQueue    queueBulk.QueueBulk[entities.ServiceLog]
	dataService interactors.DataAccess
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
	return
}

func (s *serviceLoggerImpl) persist() {
	readChan := s.logQueue.ReadAll()
	for out := range readChan {
		err := <-s.dataService.SaveServiceLogs(out.Value)
		out.IsDone(err == nil)
	}
}
