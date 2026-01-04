package main

import (
	"context"
	"errors"
	"grpc-course-protobuf/pb/chat"
	"grpc-course-protobuf/pb/common"
	"grpc-course-protobuf/pb/user"
	"io"
	"log"
	"net"
	"time"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userService struct{
	user.UnimplementedUserServiceServer
}

func (us *userService) CreateUser(ctx context.Context, userRequest *user.User) (*user.CreateResponse, error) {
	if err := protovalidate.Validate(userRequest); err != nil {
		if ve, ok := err.(*protovalidate.ValidationError); ok {
			var validations []*common.ValidationError = make([]*common.ValidationError, 0)
			for _, fieldErr := range ve.Violations{
				log.Printf("Field %s message %s ", *fieldErr.Proto.Field.Elements[0].FieldName, *fieldErr.Proto.Message)

				validations = append(validations, &common.ValidationError{
					Field: *fieldErr.Proto.Field.Elements[0].FieldName,
					Message: *fieldErr.Proto.Message,
				})
			}

			return &user.CreateResponse{
				Base: &common.BaseResponse{
					ValidationErrors: validations,
					StatusCode: 400,
					IsSuccess: false,
					Message: "validation error",
				},
			}, nil
		}
		return nil, status.Errorf(codes.InvalidArgument, "validation error %v", err)
	}

	log.Println("User is created")
	return &user.CreateResponse{
			Base: &common.BaseResponse{
				StatusCode: 200,
				IsSuccess: true,
				Message: "User created",
			},
			CreatedAt: timestamppb.Now(),
	}, nil
}

type chatService struct{
	chat.UnimplementedChatServiceServer
}

func (cs *chatService) SendMessage(stream grpc.ClientStreamingServer[chat.ChatMessage, chat.ChatResponse]) error {
	// thread infinite loop golang
	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			
			return status.Errorf(codes.Unknown, "Error receiving message %v", err)
		}

		log.Printf("Receive message: %s, to %d", req.Content, req.UserId)
	}

	return stream.SendAndClose(&chat.ChatResponse{
		Message: "Thanks for the messages!",
	})
}

func (cs *chatService) ReceiveMessage(req *chat.ReceiveMessageRequest, stream grpc.ServerStreamingServer[chat.ChatMessage]) error {
	log.Printf("Got connection request from %d\n", req.UserId)

	for i := 0; i < 10; i++ {
		err := stream.Send(&chat.ChatMessage{
			UserId: 123,
			Content: "Hi",
		})
	
		if err != nil {
			return status.Errorf(codes.Unknown, "error sending message to client %v", err)
		}
	}
	
	return nil
}

func (cs *chatService) Chat(stream grpc.BidiStreamingServer[chat.ChatMessage, chat.ChatMessage]) error {
	for {
		msg, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return status.Errorf(codes.Unknown, "error receiving message")
		}

		log.Printf("Got message from %d content: %s", msg.UserId, msg.Content)


		time.Sleep(2 * time.Second)

		err = stream.Send(&chat.ChatMessage{
			UserId: 50,
			Content: "Reply from server",
		})
		if err != nil {
			return status.Errorf(codes.Unknown, "error sending message")
		}

		err = stream.Send(&chat.ChatMessage{
			UserId: 50,
			Content: "Reply from server #2",
		})
		if err != nil {
			return status.Errorf(codes.Unknown, "error sending message")
		}
	}

	return nil
}

func main() {

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal("There is error in your net listen", err)
	}

	serv := grpc.NewServer()

	user.RegisterUserServiceServer(serv, &userService{})
	chat.RegisterChatServiceServer(serv, &chatService{})

	reflection.Register(serv)

	if err := serv.Serve(lis); err != nil {
		log.Fatal("Error running server ", err)
	}
}