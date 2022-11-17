package main

import (
	"context"
	"flag"
	"fmt"
	auctionPB "github.com/FBH93/DistributedSystemsHandIns/HandIn5/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

// Flags:
var serverName = flag.String("name", "server", "Server Name")
var port = flag.String("port", "5000", "Serer Port")

type Server struct {
	auctionPB.UnimplementedAuctionServer
	name         string
	id           int32
	port         int32
	leaderId     int32
	version      int32
	ctx          context.Context
	nodes        map[int32]auctionPB.AuctionClient
	auctionLive  bool
	highestBid   int32
	muHighestBid sync.Mutex
}

func main() {
	flag.Parse()
	s := launchServer()
	s.connectToNodes()

	// Keep server alive:
	for {
		time.Sleep(time.Second * 5)
	}
}

func launchServer() *Server {
	parsePort, _ := strconv.ParseInt(*port, 10, 32)
	ownPort := int32(parsePort)
	ctx, _ := context.WithCancel(context.Background())
	//defer cancel()
	server := &Server{
		name:  *serverName,
		port:  ownPort,
		id:    ownPort - 5000,
		ctx:   ctx,
		nodes: make(map[int32]auctionPB.AuctionClient),
	}
	log.Printf("Attemps to create listener on %d", ownPort)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port %d", ownPort)
	}

	// grpc server options:
	var opts []grpc.ServerOption

	// spin grpc server:
	grpcServer := grpc.NewServer(opts...)
	auctionPB.RegisterAuctionServer(grpcServer, server)

	go func() {
		// Serve incoming requests:
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	log.Printf("Server is listening on port %d", ownPort)
	return server
}

func (s *Server) connectToNodes() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	for i := 0; i < 3; i++ {
		port := int32(5000) + int32(i)
		if port == s.port {
			continue
		}

		// maybe timeout is useful:
		//timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
		//defer cancel()

		log.Printf("Trying to dial #%d", port)
		conn, err := grpc.Dial(fmt.Sprintf(":%d", port), opts...)
		if err != nil {
			log.Fatalf("Failed to dial on port: %d", port)
		}
		node := auctionPB.NewAuctionClient(conn)
		s.nodes[port] = node
		log.Printf("Successfully dialed and saved node: %d", port)

	}
}
