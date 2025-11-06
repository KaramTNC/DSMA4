package main

import (
	lamportclock "ChitChat/General"
	proto "ChitChat/grpc"
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
)

var users int = 0
var userIndexCounter int32 = 0
var userChannels = make(map[int32]chan proto.Message)

type ChitChatServer struct {
	proto.UnimplementedChitChatServer
}

func main() {
	server := &ChitChatServer{}

	go server.startServer()

	for {
		serverInput := readTerminal()

		if serverInput == "Quit" || serverInput == "quit" || serverInput == "Q" || serverInput == "q" {
			log.Println("[SERVER]: Shutting down server at logical time", clock.GetTime())
			return
		}
	}
}

func readTerminal() string {
	reader := bufio.NewReader(os.Stdin)

	text, _ := reader.ReadString('\n')
	text = strings.Replace(text, "\n", "", -1) // convert CRLF to LF
	text = strings.Replace(text, "\r", "", -1) // convert CRLF to LF
	clock.Iterate()
	return text
}

func (s *ChitChatServer) startServer() {
	log.Println("[SERVER]: Starting server at logical time", clock.GetTime())
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("[SERVER]: Server did not work at logical time", clock.GetTime())
	}
	clock.Iterate()
	proto.RegisterChitChatServer(grpcServer, s)

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("[SERVER]: Server did not work at logical time", clock.GetTime())
	}

	clock.Iterate()
}

func (s *ChitChatServer) GetStream(requestMessage *proto.IdMessage, stream proto.ChitChat_GetStreamServer) (err error) {
	clock.MatchTime(requestMessage.GetLamportTimestamp())

	for {
		message := <-userChannels[requestMessage.GetId()]

		message.LamportTimestamp = clock.GetTime()
		err := stream.Send(&message)
		if err != nil {
			return err
		}
	}
}

func (s *ChitChatServer) MessageToServer(ctx context.Context, msg *proto.Message) (*proto.TimeMessage, error) {
	clock.MatchTime(msg.GetLamportTimestamp())

	log.Println("[SERVER]: Received message '"+msg.GetMsg()+"' from '"+msg.GetAuthor()+"' with the id", msg.GetId(), "at logical time ", clock.GetTime())
	for _, channel := range userChannels {
		channel <- *msg
		clock.Iterate()
	}

	return &proto.TimeMessage{LamportTimestamp: clock.GetTime()}, nil
}

func (s *ChitChatServer) UserJoins(ctx context.Context, joinMessage *proto.JoinMessage) (*proto.IdMessage, error) {
	clock.MatchTime(joinMessage.GetLamportTimestamp())
	users++

	id := userIndexCounter
	userIndexCounter++

	log.Println("[SERVER]: Participant '"+joinMessage.GetUsername()+"' with id", id, "joined chit chat at logical time", clock.GetTime())

	// give the user a channel
	userChannels[id] = make(chan proto.Message, 1)

	// send message to clients that the user joined
	test := "Participant '" + joinMessage.GetUsername() + "' with id " + strconv.Itoa(int(id)) + " joined chit chat at logical time " + strconv.Itoa(int(clock.GetTime()))
	author := "server"
	Msg := proto.Message{Msg: test, Author: author, LamportTimestamp: clock.GetTime()}
	clock.Iterate()
	for _, channel := range userChannels {
		channel <- Msg
		clock.Iterate()
	}

	return &proto.IdMessage{Id: id, LamportTimestamp: clock.GetTime()}, nil
}

func (s *ChitChatServer) UserLeaves(ctx context.Context, leaveMessage *proto.LeaveMessage) (*proto.TimeMessage, error) {
	clock.MatchTime(leaveMessage.GetLamportTimestamp())
	log.Println("[SERVER]: Participant '"+leaveMessage.GetUsername()+"' with id", leaveMessage.GetId(), "left chit chat at logical time ", clock.GetTime())
	users--

	delete(userChannels, leaveMessage.GetId())

	// send message to clients that the user left
	test := "Participant '" + leaveMessage.GetUsername() + "' with id " + strconv.Itoa(int(leaveMessage.GetId())) + " left chit chat at logical time " + strconv.Itoa(int(clock.GetTime()))
	author := "server"
	Msg := proto.Message{Msg: test, Author: author, LamportTimestamp: clock.GetTime()}
	clock.Iterate()
	for _, channel := range userChannels {
		channel <- Msg
		clock.Iterate()
	}

	return &proto.TimeMessage{LamportTimestamp: clock.GetTime()}, nil
}
