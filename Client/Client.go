package main

import (
	proto "DSMA4/grpc"
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var id int32
var ch chan (string) = make(chan string)
var otherClients = []proto.UserServerClient{}

type UserServer struct {
	proto.UnimplementedUserServerServer
}

func main() {
	log.Println("Please enter a unique Id of either 1, 2 or 3:")
	id = readTerminal()
	log.Println("Current Id:", id)

	server := &UserServer{}
	go server.startServer()

	<-ch
	waitForAccept()
	connectToClients()

	for _, client := range otherClients {
		client.RequestAccess(context.Background(), &proto.Request{Id: id})
	}
}

func (s *UserServer) startServer() {
	var port = ":" + strconv.Itoa(int(8000+id))

	log.Println("Starting server on port:", port)
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	proto.RegisterUserServerServer(grpcServer, s)

	ch <- "Go ahead"
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("Server serving did not work")
	}
}

func connectToClients() {

	for i := 1; i <= 3; i++ {
		if i == int(id) {
			continue
		}

		port := ":" + strconv.Itoa(int(8000+i))
		conn, err := grpc.NewClient("localhost"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Could not connect: %v", err)
		}
		client := proto.NewUserServerClient(conn)
		otherClients = append(otherClients, client)

		log.Println("Connected to " + port)
	}

}

func (s *UserServer) RequestAccess(ctx context.Context, request *proto.Request) (*proto.TimeMessage, error) {
	log.Println("User with id: ", request.GetId(), "wanted to enter the critical section")

	return &proto.TimeMessage{}, nil
}

func readTerminal() int32 {
	var inputInt int

	for {
		var err error
		inputString := readFromUser()
		inputInt, err = strconv.Atoi(inputString)

		if err == nil {
			if inputInt < 1 || inputInt > 3 {
				log.Println("Invalid id value, must be either 1, 2 or 3")
				continue
			}

			break
		}
		log.Println("Invalid input type, please enter a valid integer")
	}

	return int32(inputInt)
}

func waitForAccept() {
	for {
		log.Println("Waiting to proceed, press Y when all servers are up")
		inputString := readFromUser()

		if inputString == "Y" {
			return
		}
	}
}

func readFromUser() string {
	reader := bufio.NewReader(os.Stdin)
	inputString, _ := reader.ReadString('\n')
	inputString = strings.TrimSuffix(inputString, "\n")
	inputString = strings.TrimSuffix(inputString, "\r")

	return inputString
}
