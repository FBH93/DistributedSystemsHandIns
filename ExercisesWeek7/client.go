package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/FBH93/DistributedSystemsHandIns/ExercisesWeek7/coordination"
	"google.golang.org/grpc"
)

type Peer struct {
	name          string
	port          string
	peers         map[string]pb.CoordinationClient // Set of clients and streams
	lampTime      int32
	mutexLampTime sync.Mutex
	mutexClient   sync.Mutex
}

// Flags:
var serverName = flag.String("name", "default peer", "Server Name")
var port = flag.String("port", "5400", "peer port")

func main() {
	fmt.Println("--- Server is starting ---")
	go launchServer()
}

func launchServer() {
	fmt.Printf("INFO: Server %s: Attempts to create listener on port %s\n", *serverName, *port)

	// Create listener lis tcp on given port or default port 5400
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err)
		return
	}
	grpcPeer := grpc.NewServer()
	peer := &peer

}
