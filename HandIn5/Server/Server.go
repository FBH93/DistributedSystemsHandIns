package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	auction "github.com/FBH93/DistributedSystemsHandIns/HandIn5/grpc"
	"google.golang.org/grpc"
)

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := &server{
		id:      ownPort,
		clients: make(map[int32]auction.AuctionClient),
		ctx:     ctx,
	}

	// Create listener tcp on port ownPort
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}
	grpcServer := grpc.NewServer()
	auction.RegisterAuctionServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	for i := 0; i < 3; i++ {
		port := int32(5000) + int32(i)

		if port == ownPort {
			continue
		}

		var conn *grpc.ClientConn
		fmt.Printf("Trying to dial: %v\n", port)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		c := auction.NewAuctionClient(conn)
		server.clients[port] = c
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		//Do something when pressing enter in console
	}
}

type server struct {
	auction.UnimplementedAuctionServer
	id      int32
	clients map[int32]auction.AuctionClient
	ctx     context.Context
}

func (s *server) bid(ctx context.Context, bid auction.Bid) (*auction.Ack, error) {

	response := &auction.Ack{Ack: true}
	return response, nil
}

func (s *server) result() {

}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
