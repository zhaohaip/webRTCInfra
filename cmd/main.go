package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"webRTCInfra/pkg/entry"
)

func main() {
	httpAddr := ":8080" // HTTP服务地址
	stunAddr := ":3478" // STUN服务地址

	server := entry.NewServer(httpAddr, stunAddr)
	if err := server.Start(); err != nil {
		log.Fatalf("failed to start server：%v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("shutting down server...")
	server.Close()
	log.Println("server closed")
}
