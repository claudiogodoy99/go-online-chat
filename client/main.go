package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"

	pb "godoy/onlinechat/proto"
)

func main() {
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())

	log.Println("Connected")
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	fmt.Print("Type your username: ")
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

	go func() {
		for {
			resp, err_n := stream.Recv()

			if err_n == io.EOF {
				log.Println("EOF")

			} else if err_n != nil {
				log.Fatalf("%v", err_n)
			}

			fmt.Println(resp.SenderUsername + " : " + resp.MessageText)

		}

	}()

	for {
		time.Sleep(1 * time.Second)

		client.SendMessage(context.Background(), &pb.ChatMessage{SenderUsername: userName, MessageText: "debug-message"})
	}
}
