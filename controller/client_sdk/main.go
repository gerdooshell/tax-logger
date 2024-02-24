package client_sdk

import (
	"context"
	"fmt"
	"github.com/gerdooshell/tax-logger/controller/client_sdk/environment"
	"github.com/gerdooshell/tax-logger/controller/client_sdk/src/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func LogError(severity, message string) (bool, error) {
	if err := environment.SetEnvironment(environment.Dev); err != nil {
		return false, err
	}
	loggerCli := newLoggerClient()
	return loggerCli.logError(context.Background(), severity, message)
}

func getLoggingServerUrl() string {
	if environment.GetEnvironment() == environment.Dev {
		return "localhost:47395"
	}
	return "tax-logger:47395"
}

type loggerClient struct {
	grpcClient logger.GRPCLoggerClient
	serverURL  string
}

var loggerClientInstance *loggerClient

func newLoggerClient() *loggerClient {
	if loggerClientInstance != nil {
		return loggerClientInstance
	}
	loggerClientInstance = &loggerClient{
		serverURL: getLoggingServerUrl(),
	}
	return loggerClientInstance
}

func (lc *loggerClient) generateDataServiceClient() error {
	if lc.grpcClient != nil {
		return nil
	}
	connection, err := grpc.Dial(lc.serverURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		err = fmt.Errorf("connection failed to the logger server")
	}
	lc.grpcClient = logger.NewGRPCLoggerClient(connection)
	//if err = connection.Close(); err != nil {
	//	return nil, fmt.Errorf("failed closing connection, error: %v\n", err)
	//}
	return err
}

func (lc *loggerClient) logError(ctx context.Context, severity, message string) (bool, error) {
	if err := lc.generateDataServiceClient(); err != nil {
		return false, err
	}
	input := &logger.SaveServiceLogRequest{
		Severity: severity,
		Message:  message,
	}
	out, err := lc.grpcClient.SaveServiceLog(ctx, input)
	success := false
	if out != nil {
		success = out.Success
	}
	return success, err
}
