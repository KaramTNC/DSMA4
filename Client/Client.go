package main

import (
	lamportclock "DSMA4/General"
	proto "DSMA4/grpc"
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var id int32
var ch chan (string) = make(chan string)
var clock lamportclock.SafeClock
var state string = "RELEASED"
var timeAtRequest int32

var queue = []int32{}

type UserServer struct {
	proto.UnimplementedUserServerServer
}

type AccessList struct {
	AccessFromOthers map[int32]bool
	sync.Mutex
}

var accessList = &AccessList{AccessFromOthers: make(map[int32]bool)}

type Clients struct {
	Clients map[int32]proto.UserServerClient
	sync.Mutex
}

var otherClients = &Clients{Clients: make(map[int32]proto.UserServerClient)}

func main() {
	log.Println("Please enter a unique Id of either 1, 2 or 3:")
	id = readTerminal()
	log.Println("Current Id:", id)

	server := &UserServer{}
	go server.startServer()

	clock.Iterate()

	<-ch

	connectToClients()
	clock.Iterate()

	for {
		// get consent from user
		waitForAcceptFromUser("Do you want to enter the critical section? press Y when ready")
		clock.Iterate()

		log.Println("Waiting for permission at logical time", clock.GetTime(), "...")
		state = "WANTED"
		timeAtRequest = clock.GetTime()
		// Asking for permission
		otherClients.Lock()
		for _, client := range otherClients.Clients {
			timeRespons, err := client.RequestAccess(context.Background(), &proto.Request{Id: id, LamportTimestamp: clock.GetTime()})
			if err != nil {
				log.Fatal("An error has accured", err)
			}
			clock.MatchTime(timeRespons.GetLamportTimestamp())
		}
		otherClients.Unlock()
		clock.Iterate()

		// waiting for permission to be granted
		waitForPermissionFromClients()

		log.Println("Access granted. Now accessing critical section at logical time", clock.GetTime())

		clock.Iterate()
		state = "HELD"
		time.Sleep(5 * time.Second) // Critical seciton
		state = "RELEASED"

		clock.Iterate()

		for _, id := range queue {
			grantAccessTo(id)
		}

		clock.Iterate()

		log.Println("now exiting the critical section at logical time", clock.GetTime())
	}
}

func (s *UserServer) RequestAccess(ctx context.Context, request *proto.Request) (*proto.TimeMessage, error) {
	clock.MatchTime(request.GetLamportTimestamp())

	var IHaveAskFirst bool = timeAtRequest < request.GetLamportTimestamp()
	var WeAskedSameTime bool = timeAtRequest == request.GetLamportTimestamp()
	var IHaveSmallestId bool = id < request.GetId()

	if state == "HELD" ||
		(state == "WANTED" && (IHaveAskFirst ||
			(WeAskedSameTime && IHaveSmallestId))) {
		queue = append(queue, request.GetId())
	} else {
		grantAccessTo(request.GetId())
	}

	return &proto.TimeMessage{LamportTimestamp: clock.GetTime()}, nil
}

func (s *UserServer) GrantAccess(ctx context.Context, response *proto.Response) (*proto.TimeMessage, error) {
	clock.MatchTime(response.GetLamportTimestamp())
	accessList.Lock()
	accessList.AccessFromOthers[response.IdFromRespondee] = true
	accessList.Unlock()
	return &proto.TimeMessage{LamportTimestamp: clock.GetTime()}, nil
}

func grantAccessTo(idToGrantAccess int32) {
	otherClients.Lock()
	timeResponse, err := otherClients.Clients[idToGrantAccess].GrantAccess(context.Background(), &proto.Response{IdFromRespondee: id, LamportTimestamp: clock.GetTime()})
	if err != nil {
		log.Fatal("An error has accured", err)
	}
	clock.MatchTime(timeResponse.GetLamportTimestamp())
	otherClients.Unlock()
}

func waitForPermissionFromClients() {
	for {
		hasAccess := true
		otherClients.Lock()
		for clientID, _ := range otherClients.Clients {
			accessList.Lock()
			if !accessList.AccessFromOthers[clientID] {
				hasAccess = false
			}
			accessList.Unlock()
		}
		otherClients.Unlock()
		if hasAccess {
			break
		}
	}

	accessList.Lock()
	for clientID, _ := range accessList.AccessFromOthers {
		accessList.AccessFromOthers[clientID] = false
	}
	accessList.Unlock()
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
		otherClients.Lock()
		otherClients.Clients[int32(i)] = client
		otherClients.Unlock()
		accessList.Lock()
		accessList.AccessFromOthers[int32(i)] = false
		accessList.Unlock()

		log.Println("Connected to " + port)
	}
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

func waitForAcceptFromUser(message string) {
	for {
		log.Println(message)
		inputString := readFromUser()

		if inputString == "Y" || inputString == "y" {
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
