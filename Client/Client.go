package main

import (
	proto "ChitChat/grpc"
	"bufio"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
)

var id int32

type UserServer struct {
	proto.UnimplementedUserServerServer
}

func main() {
	// make profile
	log.Println("Please enter a unique Id of either 1, 2 or 3:")
	id = readTerminal()
	log.Println("Current Id:", id)

	server := &UserServer{}
	go server.startServer()

	// connecting to server
	//log.Println(" Connecting to server ...")
	//conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	//if err != nil {
	//log.Fatalf("[CLIENT]: Could not connect: %v", err)
	//}
	//client = proto.NewChitChatClient(conn)
	//joinServer()

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

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("Server serving did not work")
	}
}

func messageServer(message string) {
	//_, err := client.MessageToServer(context.Background(), &proto.Message{Msg: message, Id: id})
	//if err != nil {
	//	log.Fatalf("[CLIENT]: Could not send message: %v", err)
	//}
}

func printMessage(msg *proto.Message) {
	log.Println(msg.Author + ": " + msg.Msg)
}

func readTerminal() int32 {
	var inputInt int

	for {
		reader := bufio.NewReader(os.Stdin)
		var err error
		inputString, _ := reader.ReadString('\n')
		inputString = strings.TrimSuffix(inputString, "\n")
		inputString = strings.TrimSuffix(inputString, "\r")

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
