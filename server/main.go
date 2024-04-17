package main

import (
	"context"
	"fmt"
	pb "godoy/onlinechat/proto"
	"log"
	"net"

	"google.golang.org/grpc"
)

type server struct {
	pb.ChatServiceServer
	messages []chan message
}

type user struct {
	username  string
	channelId int
}

type message struct {
	message  string
	userName string
}

func (s *server) UpdateUserStatus(in *pb.UserStatusChange, srv pb.ChatService_UpdateUserStatusServer) error {

	if in.NewStatus == pb.UserStatusChange_ONLINE {
		ch := make(chan message)

		s.messages = append(s.messages, ch)
		index := len(s.messages) - 1

		log.Println("Creating new channel: ", index)

		for m := range s.messages[index] {
			log.Println("Sending message to client")
			mes := pb.ChatMessage{
				SenderUsername: m.userName,
				MessageText:    m.message,
			}
			srv.Send(&mes)
		}
	} else {
		srv.Context().Done()
		return nil
	}

	return nil
}

func (s *server) SendMessage(ctx context.Context, msg *pb.ChatMessage) (*pb.ChatMessageResponse, error) {
	log.Println("Recieved the message, sending to all channels")

	for i := range s.messages {
		fmt.Println("sending to: ", i)
		s.messages[i] <- message{
			message:  msg.MessageText,
			userName: msg.SenderUsername,
		}

	}

	return &pb.ChatMessageResponse{
		Ok:  true,
		Err: "",
	}, nil

}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	maxUsers := 0
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()

	pb.RegisterChatServiceServer(s, &server{
		messages: make([]chan message, maxUsers),
	})

	sErr := s.Serve(lis)

	if sErr != nil {
		log.Fatal("failed to serve: %v", err)
	} else {
		log.Println("Server up and running")
	}

}
