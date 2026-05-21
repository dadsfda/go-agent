package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"smart-recruitment/logic-grpc-service/internal/app"
	"smart-recruitment/logic-grpc-service/internal/pbserver"
	logicpb "smart-recruitment/logic-grpc-service/proto"
)

func main() {
	addr := getenv("LOGIC_ADDR", ":9001")
	service, err := app.NewServiceFromEnv(os.Getenv("MYSQL_DSN"))
	if err != nil {
		log.Fatalf("初始化业务服务失败: %v", err)
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("启动 Logic 监听失败: %v", err)
	}
	server := grpc.NewServer()
	logicpb.RegisterLogicServiceServer(server, pbserver.New(service))
	log.Printf("Logic gRPC 服务已启动: %s", addr)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Logic 服务退出: %v", err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
