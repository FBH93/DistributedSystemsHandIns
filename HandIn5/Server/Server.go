package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	auction "github.com/FBH93/DistributedSystemsHandIns/HandIn5/grpc"
	"google.golang.org/grpc"
)

type server struct {
	auction.UnimplementedAuctionServer
	id            int32
	isLeader      bool
	servers       map[int32]auction.AuctionClient
	ctx           context.Context
	ongoing       bool
	highestBid    int32
	highestBidder int32
	lock          sync.Mutex
}

func main() {
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := &server{
		id:            ownPort,
		isLeader:      false,
		servers:       make(map[int32]auction.AuctionClient),
		ctx:           ctx,
		ongoing:       true,
		highestBid:    0,
		highestBidder: 0,
	}

	// Create listener tcp on localhost:ownPort
	// (listening on localhost explicitly avoid Windows firewall confirmation upon doing go run)
	list, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}
	fmt.Printf("Now listening on port %v \n", ownPort)
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
		server.servers[port] = c
	}
	fmt.Printf("Successfully dailed all other servers, normal operation commencing...\n")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		//Do something when pressing enter in console
	}
}

func (s *server) Bid(ctx context.Context, bid *auction.Bid) (*auction.Ack, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	fmt.Printf("Received bid from client %v for value %v. ", bid.Id, bid.Bid)
	if !s.isLeader {
		s.isLeader = true // set isLeader if client sends Bid to server directly
		fmt.Printf("Now considered leader, as bid from client was recieved")
	}
	if bid.Bid > s.highestBid {
		s.highestBid = bid.Bid
		s.highestBidder = bid.Id
		fmt.Printf("Bid was higher than previously top bid, so updating leader to client %v", bid.Id)
	}
	fmt.Printf("\n")
	s.PropagateBid(bid) // forward bid to all other servers
	response := &auction.Ack{Ack: true}
	return response, nil
}

func (s *server) Result(ctx context.Context, req *auction.ResultReq) (*auction.Outcome, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	response := &auction.Outcome{
		Ongoing:       s.ongoing,
		HighestBid:    s.highestBid,
		HighestBidder: s.highestBidder,
	}
	return response, nil
}

// PropagateBid propagates an incoming bid to all other servers, to keep their states in sync
func (s *server) PropagateBid(bid *auction.Bid) {
	timeOutCtx, cancel := context.WithTimeout(s.ctx, time.Second*15)
	defer cancel()
	for id, server := range s.servers {
		_, err := server.RelayBid(timeOutCtx, bid)
		if err != nil {
			fmt.Printf("Error: No response from server on port %v, now assumed to be dead", id)
			delete(s.servers, id)
		}
	}
}

// RelayBid accepts a bid sent from a fellow server and applies the bid, *without* subsequently relaying to other servers
func (s *server) RelayBid(ctx context.Context, bid *auction.Bid) (*auction.Ack, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.isLeader {
		s.isLeader = false // set isLeader to false if other server relays Bid to this server
		fmt.Printf("Stepping down as leader, as relayed bid was recieved")
	}
	fmt.Printf("Received relayed bid from client %v for value %v. ", bid.Id, bid.Bid)
	if bid.Bid > s.highestBid {
		s.highestBid = bid.Bid
		s.highestBidder = bid.Id
		fmt.Printf("Bid was higher than previously top bid, so updating leader to client %v", bid.Id)
	}
	fmt.Printf("\n")
	response := &auction.Ack{Ack: true}
	return response, nil
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
