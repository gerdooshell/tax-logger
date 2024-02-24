package server

import (
	"fmt"
	"github.com/gerdooshell/tax-logger/controller/protobuf/src/logger"
	"google.golang.org/grpc"
	"net"
)

func ServeGRPC() error {
	port := 47395
	ip := "0.0.0.0"
	address := fmt.Sprintf("%v:%v", ip, port)
	grpcServer := grpc.NewServer()
	loggerSrvHandler := NewLoggerServerHandler()
	logger.RegisterGRPCLoggerServer(grpcServer, loggerSrvHandler)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	fmt.Println("Staring GRPC Server")
	if err = grpcServer.Serve(listener); err != nil {
		return err
	}
	return nil
}
