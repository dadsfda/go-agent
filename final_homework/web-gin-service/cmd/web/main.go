package main

import (
	"log"
	"os"

	"smart-recruitment/web-gin-service/internal/httpapi"
	"smart-recruitment/web-gin-service/internal/rpcclient"
)

func main() {
	logicAddr := getenv("LOGIC_ADDR", "127.0.0.1:9001")
	webAddr := getenv("WEB_ADDR", ":8080")
	client, err := rpcclient.Dial(logicAddr)
	if err != nil {
		log.Fatalf("连接 Logic 服务失败: %v", err)
	}
	defer client.Close()
	log.Printf("Web 服务连接 Logic: %s", logicAddr)
	if err := httpapi.NewRouter(client).Run(webAddr); err != nil {
		log.Fatalf("Web 服务退出: %v", err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
