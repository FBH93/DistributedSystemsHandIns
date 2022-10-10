package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/FBH93/DistributedSystemsHandIns/HandIn3/ChittyChat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

// Flags:
var clientName = flag.String("name", "Default Client", "Name of client")
var serverPort = flag.String("port", "5400", "tcp server")

var server pb.ChittyChatClient
var serverConn *grpc.ClientConn

func main() {
	flag.Parse()

	fmt.Println("Attempting to connect to server")
	connectToServer()

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
	log.Printf("The connection is: %s\n", conn.GetState().String())
}
