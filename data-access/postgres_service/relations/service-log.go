package relations

import (
	"github.com/gerdooshell/tax-logger/entities"
	"time"
)

type ServiceLogModel struct {
	Id          int       `gorm:"<-;primaryKey"`
	Timestamp   time.Time `gorm:"<-;not null"`
	Severity    string    `gorm:"<-;not null"`
	Message     string    `gorm:"<-;not null"`
	ServiceName string    `gorm:"<-;not null"`
	StackTrace  string    `gorm:"<-;"`
	ProcessId   string    `gorm:"<-;"`
}

func (sl *ServiceLogModel) TableName() string {
	return "logs.service_log"
}

func NewServiceLogModelFromEntity(entity entities.ServiceLog) ServiceLogModel {
	return ServiceLogModel{
		Severity:    entity.Severity.ToString(),
		Timestamp:   entity.Timestamp,
		Message:     entity.Message,
		ServiceName: entity.Origin.ServiceName.ToString(),
		StackTrace:  entity.Origin.StackTrace,
		ProcessId:   entity.Origin.ProcessId,
	}
}
