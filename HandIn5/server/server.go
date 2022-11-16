package main

import (
	"flag"
	"fmt"
	auctionPB "github.com/FBH93/DistributedSystemsHandIns/HandIn5/grpc"
	"google.golang.org/grpc"
	"log"
	"net"
	"strconv"
	"sync"
)

// Flags:
var serverName = flag.String("name", "server", "Server Name")
var port = flag.String("port", "5000", "Serer Port")

type Server struct {
	auctionPB.UnimplementedAuctionServer
	name         string
	id           int32
	port         int32
	leader       bool
	version      int32
	nodes        map[int32]auctionPB.AuctionServer
	auctionLive  bool
	highestBid   int32
	muHighestBid sync.Mutex
}

func main() {
	flag.Parse()
	launchServer()

}

func launchServer() {
	parsePort, _ := strconv.ParseInt(*port, 10, 32)
	ownPort := int32(parsePort)
	server := &Server{
		name:  *serverName,
		port:  ownPort,
		id:    ownPort - 5000,
		nodes: make(map[int32]auctionPB.AuctionServer),
	}
	log.Printf("Attemps to create listener on %d", ownPort)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d", ownPort)
	}

	// grpc server options:
	var opts []grpc.ServerOption

	// spin grpc server:
	grpcServer := grpc.NewServer(opts...)
	auctionPB.RegisterAuctionServer(grpcServer, server)

	// serve incoming requests:
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
	//go func() {
	//	// Serve incoming requests:
	//	if err := grpcServer.Serve(lis); err != nil {
	//		log.Fatalf("Failed to serve: %v", err)
	//	}
	//}()
}

func (s *Server) connectToNodes() {
	for i := 0; i < 3; i++ {
		port := int32(5000) + int32(i)
		if port == s.port {
			continue
		}

	}
}
