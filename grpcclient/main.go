package main

import (
	"context"
	"grpc-course-protobuf/pb/user"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	clientConn, err := grpc.NewClient("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to create client ", err)
	}

	userClient := user.NewUserServiceClient(clientConn)
	res, err := userClient.CreateUser(context.Background(), &user.User{
		Age: 30,
	})
	if err != nil {
		st, ok := status.FromError(err)
		// Error grpc
		if ok {
			if st.Code() == codes.InvalidArgument {
				log.Println("There is InvalidArgument error: ", st.Message())
			} else if st.Code() == codes.Unknown {
				log.Println("There is Unknown error: ", st.Message())
			} else if st.Code() == codes.Internal {
				log.Println("There is Internal error: ", st.Message())
			}

			return
		}
		
		log.Println("Failed to send message ", err)
		return
	}

	log.Println("Response from server ", res.Message)
}