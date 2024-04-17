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

	pb "godoy/onlinechat/proto"

	"github.com/jroimartin/gocui"
	"google.golang.org/grpc"
)

var (
	g            *gocui.Gui
	inputView    *gocui.View
	outputView   *gocui.View
	receivedMsgs []string
	userName     string
	client       pb.ChatServiceClient
	// messages     chan string
)

func main() {
	// messages = make(chan string)

	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	client = pb.NewChatServiceClient(conn)

	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	fmt.Print("Type your username: ")
	reader := bufio.NewReader(os.Stdin)
	userName, err = reader.ReadString('\n')
	userName = strings.TrimSpace(userName)

	if err != nil {
		log.Fatal(err)
	}

	g, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true
	g.Highlight = true

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, sendMsg); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	stream, err := client.UpdateUserStatus(context.Background(), &pb.UserStatusChange{
		Username:  userName,
		NewStatus: pb.UserStatusChange_ONLINE,
	})

	if err != nil {
		log.Panicln(err)
	}

	go func() {
		for {

			resp, err := stream.Recv()

			if err == io.EOF {
				log.Fatal("EOF")
				// TODO RECONNECT

			} else if err != nil {
				log.Fatal("err")
				// TODO RECONNECT
			}

			msg := resp.SenderUsername + " : " + resp.MessageText
			receivedMsgs = append(receivedMsgs, msg)
		}
	}()

	displayTicker := time.NewTicker(100 * time.Millisecond)

	go func() {
		for range displayTicker.C {
			g.Update(func(g *gocui.Gui) error {
				updateOutputView()
				return nil
			})
		}
	}()

	g.MainLoop()

}

func updateOutputView() {
	outputView.Clear()
	for _, msg := range receivedMsgs {
		fmt.Fprintln(outputView, msg)
	}
}

func sendMsg(g *gocui.Gui, v *gocui.View) error {
	msg := strings.TrimSpace(inputView.Buffer())
	if msg != "" {

		client.SendMessage(context.Background(), &pb.ChatMessage{
			SenderUsername: userName,
			MessageText:    msg,
		})

		if err := inputView.SetCursor(0, 0); err != nil {
			return err
		}

		inputView.Clear()
	}

	return nil
}
