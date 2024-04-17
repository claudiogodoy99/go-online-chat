package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"

	pb "godoy/onlinechat/proto"
)

func main() {
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	reader := bufio.NewReader(os.Stdin)
	userName, err := reader.ReadString('\n')

	userName = strings.TrimSpace(userName)

	if err != nil {
		log.Fatal(err)
	}

	client := pb.NewChatServiceClient(conn)

	in := &pb.UserStatusChange{
		Username:  userName,
		NewStatus: pb.UserStatusChange_ONLINE,
	}

	stream, err := client.UpdateUserStatus(context.Background(), in)

	if err != nil {
		log.Fatal("Deu ruim")
	}

	// done := make(chan bool)

	go func() {
		for {
			resp, err_n := stream.Recv()

			if err_n == io.EOF {
				// log.Println("EOF")

			} else if err_n != nil {
				// log.Println("cannot receive %v", err)
			}

			if resp.SenderUsername != "" && resp.SenderUsername != userName {
				fmt.Println(resp.SenderUsername + " : " + resp.MessageText)
			}

		}

	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil {
			log.Fatal(err)
		}
		client.SendMessage(context.Background(), &pb.ChatMessage{SenderUsername: userName, MessageText: line})
	}
}
