package main

import (
	"context"
	"errors"
	"fmt"
	pb "godoy/onlinechat/proto"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type server struct {
	pb.ChatServiceServer
	messages            []chan message
	messagesChannelSlot []bool
	userMap             map[string]int
	sync                sync.Mutex
}

type message struct {
	message  string
	userName string
}

func main() {

	lis, err := net.Listen("tcp", ":50051")
	maxUsers := 2
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()

	pb.RegisterChatServiceServer(s, &server{
		messages:            make([]chan message, maxUsers),
		messagesChannelSlot: make([]bool, maxUsers),
		userMap:             make(map[string]int),
		sync:                sync.Mutex{},
	})

	err = s.Serve(lis)

	if err != nil {
		log.Fatal("failed to serve")
	} else {
		log.Println("Server up and running")
	}

}

func (s *server) UpdateUserStatus(in *pb.UserStatusChange, srv pb.ChatService_UpdateUserStatusServer) error {

	if in.NewStatus == pb.UserStatusChange_ONLINE {

		index, err := s.addNewUser(in.Username)

		if err != nil {
			log.Println(err.Error())
			return err
		}

		for {
			select {
			case <-srv.Context().Done():
				{
					log.Println("Context done, deleting user and releasing resources")
					s.deleteUser(in.Username)
					return nil
				}
			case m := <-s.messages[index]:
				{
					log.Println("Sending message to client")
					mes := pb.ChatMessage{
						SenderUsername: m.userName,
						MessageText:    m.message,
					}
					srv.Send(&mes)
				}
			}
		}
	} else {
		return nil
	}

}

func (s *server) addNewUser(userName string) (int, error) {
	s.sync.Lock()
	defer s.sync.Unlock()

	if _, ok := s.userMap[userName]; ok {
		return 0, errors.New("user exists")
	}

	slot := 0
	found := false
	for i := range s.messagesChannelSlot {
		if !s.messagesChannelSlot[i] {
			found = true
			slot = i
			break
		}
	}

	if !found {
		return 0, errors.New("lack slot")
	}

	ch := make(chan message)

	s.messages[slot] = ch

	s.messagesChannelSlot[slot] = true
	s.userMap[userName] = slot

	return slot, nil
}

func (s *server) deleteUser(userName string) error {
	s.sync.Lock()
	defer s.sync.Unlock()

	if i, ok := s.userMap[userName]; ok {
		delete(s.userMap, userName)
		s.messagesChannelSlot[i] = false
		oldCn := s.messages[i]

		defer close(oldCn)
	} else {
		return errors.New("not found")
	}

	return nil
}

func (s *server) SendMessage(ctx context.Context, msg *pb.ChatMessage) (*pb.ChatMessageResponse, error) {
	log.Println("Received the message, sending to all channels")
	s.sync.Lock()
	for _, i := range s.userMap {
		fmt.Println("sending to: ", i)
		if s.messages[i] != nil && s.messagesChannelSlot[i] {
			s.messages[i] <- message{
				message:  msg.MessageText,
				userName: msg.SenderUsername,
			}
		}
	}
	s.sync.Unlock()

	return &pb.ChatMessageResponse{
		Ok:  true,
		Err: "",
	}, nil

}
