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
	log.Printf("Starting peer on port %v", ownPort)
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
		log.Printf("Trying to dial: %v\n", port)
		// unsuccessful dialing attempts become blocking by grpc.WithBlock(), as to give us time to launch other peers
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		// add dialed node to clients map
		c := ping.NewPingClient(conn)
		p.clients[port] = c
	}

	// Ask all nodes for permission to enter critical section, and execute critical section
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		p.enterCritical()
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
	log.Printf("Recieved request from %v to enter critical section", req.Id)
	retryAttempts := 0 //To handle deadlocks, we count how many times peer has tried to check the state

	for { // busy wait / infinite loop
		if !p.held && !p.wanted { // respond iff lock is not wanted nor held (i.e. the peer is busy entering / in the critical section):
			log.Printf("Sending go-ahead to node %v's request to enter critical section\n", req.Id)
			rep := &ping.Reply{Id: p.id, Permission: true}
			return rep, nil
		} else { // Waiting position, keep looping until lock is not wanted nor held.
			if retryAttempts > 10 {
				rep := &ping.Reply{Id: p.id, Permission: false}
				return rep, nil //Return reply with permission denied if 10 tries has passed with no progress to prevent deadlocks
			}
			retryAttempts++
			time.Sleep(time.Second * 1) //Wait 1 second before trying again.
			continue
		}
	}

	// To avoid compiler's static checking for ensuring a return statement.
	// This should never be hit...
	log.Printf("Something went very wrong. Check if locks are working as intended.")
	return nil, nil
}

// Function to send ping to all peers by iterating over a map of clients.
// Returns true iff all other peers reponds positively
// Returns false if any requests to peers times out (indicating dead peer or deadlock)
func (p *peer) sendPingToAll() bool {
	request := &ping.Request{Id: p.id}
	for id, client := range p.clients {
		reply, err := client.Ping(p.ctx, request)
		if err != nil {
			fmt.Println("something went wrong\n")
			return false
		}
		if !reply.Permission {
			return false
		}
		log.Printf("Got positive reply from id %v: %v\n", id, reply.Id)
	}
	return true
}

// Peer tries to enter critical section
func (p *peer) enterCritical() {
	p.wanted = true //prevents peer from responding while lock is wanted.
	log.Printf("%v wants to enter the critical section \n", p.id)
	if p.sendPingToAll() { //Wait for response from all peers before proceeding. Proceed iff the method returns true (i.e. all other peers responded)
		p.held = true
		p.wanted = false
		log.Printf("%v has acquired the lock for the critical section", p.id)
		p.doCritical() //Enter critical section
		p.held = false
		log.Printf("%v has released the lock for the critical section", p.id)
	} else {
		p.wanted = false
		log.Printf("%v did not get a response from all clients. Will request to enter critical section again shortly.", p.id)
		time.Sleep(time.Second * (time.Duration(p.id) % 5000)) //Wait a dynamic amount of time, to avoid entering deadlock again. To avoid 2 peers retrying after the same delay and deadlocking again.
		p.enterCritical()                                      //Try to enter critical again.
	}
}

// function to emulate a critical section, in our case simply logging an event
func (p *peer) doCritical() {
	time.Sleep(time.Second * 4) //Simulate the time it takes to do something critical
	log.Printf("%v ran critical section - Yay \n", p.id)
}
