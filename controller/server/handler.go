package server

import (
	"context"
	"fmt"
	"time"

	"github.com/gerdooshell/tax-communication/src/logger"
	"github.com/gerdooshell/tax-logger/entities"
	serviceName "github.com/gerdooshell/tax-logger/entities/constants/service-name"
	"github.com/gerdooshell/tax-logger/entities/severity"
	serviceLogger "github.com/gerdooshell/tax-logger/interactors/service-logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type LoggerServerHandler interface {
	logger.GRPCLoggerServer
}

func NewLoggerServerHandler() LoggerServerHandler {
	return &loggerServerHandler{}
}

type loggerServerHandler struct {
}

func (l *loggerServerHandler) SaveServiceLog(ctx context.Context, request *logger.SaveServiceLogReq) (empty *emptypb.Empty, err error) {
	//apiKey := request.GetAPIKey()  TODO: validate api key
	loggerService := serviceLogger.GetServiceLoggerInstance()
	severityValue, err := severity.FromString(request.GetSeverity())
	if err != nil {
		return
	}
	origin := request.GetOriginLog()
	srvName, err := serviceName.FromString(origin.GetServiceName())
	if err != nil {
		return
	}
	fmt.Println(request.GetTimestamp())
	serviceLog := entities.ServiceLog{
		Timestamp: time.Now(),
		Severity:  severityValue,
		Message:   request.GetMessage(),
		Origin: entities.OriginLog{
			ProcessId:   origin.GetProcessId(),
			ServiceName: srvName,
			StackTrace:  origin.GetStackTrace(),
		},
	}
	err = loggerService.Log(serviceLog)
	return
}
