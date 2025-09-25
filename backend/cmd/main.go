package main

import (
    "log"
	
    rabbit "backend/internal/infrastructure"
)

func main() {
    amqpURI := "amqp://guest:guest@localhost:5672/"


    publisher, err := rabbit.GetRPCconnection(amqpURI)
    if err != nil {
        log.Fatalf("Failed to initialize RPC publisher: %v", err)
    }


    // Отправка RPC задачи и ожидание ответа
    response, err := publisher.PublishNmap(req)
    if err != nil {
        log.Fatalf("Failed to publish nmap task: %v", err)
    }


}
