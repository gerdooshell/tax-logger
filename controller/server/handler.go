package server

import (
	"context"
	"github.com/gerdooshell/tax-logger/controller/protobuf/src/logger"
	"github.com/gerdooshell/tax-logger/entities"
	"github.com/gerdooshell/tax-logger/entities/severity"
	serviceLogger "github.com/gerdooshell/tax-logger/interactors/service-logger"
	"time"
)

type LoggerServerHandler interface {
	logger.GRPCLoggerServer
}

func NewLoggerServerHandler() LoggerServerHandler {
	return &loggerServerHandler{}
}

type loggerServerHandler struct {
}

func (l *loggerServerHandler) SaveServiceLog(ctx context.Context, request *logger.SaveServiceLogRequest) (*logger.SaveServiceLogResponse, error) {
	loggerService := serviceLogger.GetServiceLoggerInstance()
	severityValue, err := severity.FromString(request.GetSeverity())
	if err != nil {
		return nil, err
	}
	serviceLog := entities.ServiceLog{
		Timestamp: time.Now(),
		Severity:  severityValue,
		Message:   request.Message,
	}
	err = loggerService.Log(serviceLog)
	return &logger.SaveServiceLogResponse{Success: err == nil}, err
}
