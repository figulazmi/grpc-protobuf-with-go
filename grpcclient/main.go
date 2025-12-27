package main

import (
	"context"
	"grpc-course-protobuf/pb/chat"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	clientConn, err := grpc.NewClient("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to create client ", err)
	}

	chatClient := chat.NewChatServiceClient(clientConn)
	stream, err := chatClient.SendMessage(context.Background())
	if err != nil {
		log.Fatal("Failed to send message", err)
	}

	err = stream.Send(&chat.ChatMessage{
		UserId: 123,
		Content: "Hello from client",
	})
	if err != nil {
		log.Fatal("Failed to send via stream ", err)
	}
	
	err = stream.Send(&chat.ChatMessage{
		UserId: 123,
		Content: "Hello again",
	})
	if err != nil {
		log.Fatal("Failed to send via stream ", err)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("Failed close", err)
	}
	log.Println("Connection is closed. Message: ", res.Message)
}