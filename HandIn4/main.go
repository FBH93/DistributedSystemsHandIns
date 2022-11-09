package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	ping "github.com/FBH93/DistributedSystemsHandIns/HandIn4/grpc"
	"google.golang.org/grpc"
)

//Run file in the folder by command "go run . X" where X is a number

func main() {
	//Set port based on input when running the go file.
	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 5000

	//Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Create the peer as a server
	p := &peer{
		id:      ownPort,
		clients: make(map[int32]ping.PingClient),
		ctx:     ctx,
		held:    false,
		wanted:  false,
	}

	// Create listener tcp on port ownPort
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}
	grpcServer := grpc.NewServer()
	ping.RegisterPingServer(grpcServer, p)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	// establish connection to all peers (harccoded to 3 total peers in this case)
	for i := 0; i < 3; i++ {
		// ports of peers must be in direct succession from port 5000, i.e. 5000, 5001, 5002, and so on
		port := int32(5000) + int32(i)

		// skip dialing to one self
		if port == ownPort {
			continue
		}

		var conn *grpc.ClientConn
		fmt.Printf("Trying to dial: %v\n", port)
		// unsuccessful dialing attempts become blocking by grpc.WithBlock(), as to give us time to launch other nodes
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		// add dialed node to clients map
		c := ping.NewPingClient(conn)
		p.clients[port] = c
	}

	//Send a ping to all when anything is input in the console.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		p.sendPingToAll()
	}
}

type peer struct {
	ping.UnimplementedPingServer
	id      int32
	clients map[int32]ping.PingClient
	ctx     context.Context
	held    bool
	wanted  bool
}

// Function to receive and respond to a ping from a peer.
func (p *peer) Ping(ctx context.Context, req *ping.Request) (*ping.Reply, error) {
	rep := &ping.Reply{Id: p.id}

	for true {
		if p.held == false && p.wanted == false { // respond when lock is not wanted / held (i.e. the peer is busy entering / in the critical section):
			return rep, nil
		} else { // Waiting position, keep looping until lock is not wanted or held.
			continue
		}
	}

	// To avoid compiler's static checking for ensuring a return statement.
	// This should never be hit...
	log.Printf("Something went very wrong. Check if locks are working as intended.")
	return nil, nil
}

// Function to send ping to all peers by iterating over a map of clients.
func (p *peer) sendPingToAll() {
	request := &ping.Request{Id: p.id}
	for id, client := range p.clients {
		reply, err := client.Ping(p.ctx, request)
		if err != nil {
			fmt.Println("something went wrong")
		}
		fmt.Printf("Got reply from id %v: %v\n", id, reply.Id)
	}
}

func (p *peer) requestLock() {
	p.wanted = true   //prevents peer from responding while lock is wanted.
	p.sendPingToAll() //Wait for response from all peers before proceeding.
	p.held = true
	p.wanted = false
	p.doCritical()
	p.held = false
}

// function to emulate a critical section, in our case simply logging an event
func (p *peer) doCritical() {
	time.Sleep(1000) //Simulate the time it takes to do something critical
	log.Printf("%v ran critical section - Yay \n", p.id)
}
