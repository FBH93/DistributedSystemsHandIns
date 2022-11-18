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
var port = flag.String("port", "5400", "Serer Port")

type Server struct {
	auctionPB.UnimplementedAuctionServer
	name         string
	id           int32
	port         int32
	leaderId     int32
	leader       bool
	version      int32
	ctx          context.Context
	nodes        map[int32]auctionPB.Nodes_ConnectNodesServer // map over nodes and streams
	clients      map[int32]auctionPB.AuctionClient
	auctionLive  bool
	highestBid   int32
	muHighestBid sync.Mutex
}

func main() {
	flag.Parse()

	parsePort, _ := strconv.ParseInt(*port, 10, 32)
	ownPort := int32(parsePort)
	ctx, _ := context.WithCancel(context.Background())
	//defer cancel()
	s := &Server{
		name: *serverName,
		port: ownPort,
		id:   ownPort - 5400,
		ctx:  ctx,
		//nodes: make(map[int32]auctionPB.AuctionClient),
	}
	if s.port == 5400 {
		s.leader = true
		s.launchServer()
	} else {
		s.leader = false
		leader := s.connectToLeader()
		stream, err := leader.ConnectNodes(s.ctx)
		go receive(stream)
		if err != nil {
			log.Printf("Error getting stream from leader")
		}

	}

	// Keep server alive:
	for {
		time.Sleep(time.Second * 5)
	}
}

// TODO Implement ConnectNodes
// Should be renamed to 'Update'?
func (s *Server) ConnectNodes(nodeStream auctionPB.Nodes_ConnectNodesServer) error {
	ping, err := nodeStream.Recv()
	if err != nil {
		// Stream closed
	}
	s.highestBid = ping.HighestBid

	return nil
}

// TODO Implement receive
func receive(stream auctionPB.Nodes_ConnectNodesClient) {

}

func (s *Server) launchServer() {

	log.Printf("Attemps to create listener on %d", s.port)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d", s.port)
	}

	// grpc server options:
	var opts []grpc.ServerOption

	// spin grpc server:
	grpcServer := grpc.NewServer(opts...)
	auctionPB.RegisterAuctionServer(grpcServer, s)

	go func() {
		// Serve incoming requests:
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	log.Printf("Server is listening on port %d", s)
}

func (s *Server) connectToLeader() auctionPB.NodesClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	// maybe timeout is useful:
	//timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()

	log.Printf("Trying to dial #%d", port)
	conn, err := grpc.Dial(fmt.Sprintf(":%d", port), opts...)
	if err != nil {
		log.Fatalf("Failed to dial on port: %d", port)
	}
	leader := auctionPB.NewNodesClient(conn)
	log.Printf("Successfully connected to the leader")
	return leader
}
