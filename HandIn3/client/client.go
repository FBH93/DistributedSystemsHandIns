package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/FBH93/DistributedSystemsHandIns/HandIn3/ChittyChat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Flags:
var clientName = flag.String("name", "Default Client", "Name of client")
var serverPort = flag.String("port", "5400", "tcp server")

var lampTime int32 = 0 //client starts with lamport time 0

var server pb.ChittyChatClient
var serverConn *grpc.ClientConn

func main() {
	flag.Parse()

	fmt.Println("Attempting to connect to server")
	connectToServer()

	// Get stream from server
	stream, err := server.Chat(context.Background())
	if err != nil {
		log.Println(err)
		log.Printf("Could not connect to server. Is it running?")
		return
	}

	go receive(stream)

	// Ensure first message to server initializes the welcome message
	if err := stream.Send(&pb.ChatRequest{Msg: "", ClientName: *clientName}); err != nil {
		log.Fatal(err)
	}
	parseInput(stream)

	//TODO: Leave message
}

func increaseLamptime(receivedTime int32) {
	if lampTime > receivedTime {
		lampTime++
	} else {
		lampTime = receivedTime + 1
	}
}

// Receive and print stream from server
func receive(stream pb.ChittyChat_ChatClient) {
	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		increaseLamptime(resp.Time) //before printing received msg, increase time
		log.Printf("[T:%d]Received: %s", lampTime, resp.Msg)
	}
}

func connectToServer() {
	// Dial options:
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	// Time out on the connection
	timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Dial the server to get a connection:
	log.Printf("Client %s: Attempts to dial on port %s\n", *clientName, *serverPort)
	conn, err := grpc.DialContext(timeContext, fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Failed to dial: %v\n", err)
		return
	}

	server = pb.NewChittyChatClient(conn)
	serverConn = conn
	increaseLamptime(lampTime) //Client has connected to server, increase time.
	log.Printf("[T:%d] The connection is: %s\n", lampTime, conn.GetState().String())
}

func parseInput(stream pb.ChittyChat_ChatClient) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Post a message to ChittyChat:")

	for {
		fmt.Printf("-> ")

		// Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Input gave an error: %v", err)
		}
		// Trim input
		input = strings.TrimSpace(input)

		if serverConn.GetState().String() != "READY" {
			//TODO: Try to substitute with log.Fatalf()
			log.Printf("Client %s: Something was wrong with the connection to the server :(", *clientName)
			continue
		}

		//TODO: Event logic goes here:
		prefix := *clientName + ": "
		increaseLamptime(lampTime) //Before sending message to server, increase lamptime
		if err := stream.Send(&pb.ChatRequest{Msg: prefix + input, ClientName: *clientName, Time: lampTime}); err != nil {
			log.Fatal(err)
		}
	}
}
