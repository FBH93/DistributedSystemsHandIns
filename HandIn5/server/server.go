package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	auctionPB "github.com/FBH93/DistributedSystemsHandIns/HandIn5/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Flags:
var serverName = flag.String("name", "server", "Server Name")
var port = flag.String("port", "5400", "Serer Port")

type Server struct {
	auctionPB.UnimplementedNodesServer
	auctionPB.UnimplementedAuctionServer
	name        string
	id          int32
	port        int32
	leaderId    int32
	leader      bool
	version     int32 // Should version have a lock?
	crashes     int32
	ctx         context.Context
	nodes       map[int32]auctionPB.Nodes_UpdateNodesServer // map over nodes and streams
	leaderQueue []int32                                     // Queue of potential leaders
	//clients     map[string]auctionPB.AuctionClient
	auctionLive  bool
	highBidder   int32
	highBid      int32
	muHighBid    sync.Mutex
	muHighBidder sync.Mutex
}

func main() {
	flag.Parse()

	parsePort, _ := strconv.ParseInt(*port, 10, 32)
	ownPort := int32(parsePort)
	ctx, _ := context.WithCancel(context.Background())
	//defer cancel()
	s := &Server{
		name:        *serverName,
		port:        ownPort,
		id:          ownPort - 5400,
		ctx:         ctx,
		nodes:       make(map[int32]auctionPB.Nodes_UpdateNodesServer),
		highBid:     0,
		leaderId:    0,
		crashes:     0,
		leaderQueue: []int32{},
	}

	//COMMENT OUT THESE TWO LINES TO REMOVE LOGGING TO TXT
	//logfile := setLog() //print log to a log.txt file instead of the console
	//defer logfile.Close()

	// make server listening on port 5400 the first leader when starting program
	if s.port == 5400 {
		s.leader = true
		s.auctionLive = false
		s.launchServer()
	} else {
		s.leader = false
		connectionToLeader(s)

	}

	// Keep server alive:
	for {
		time.Sleep(time.Second * 5)
	}
}

func connectionToLeader(s *Server) {
	leader := s.connectToLeader()
	stream, err := leader.UpdateNodes(context.Background())
	if err != nil {
		log.Printf("Error getting stream from leader")
	}

	// Send join statement to leader:
	if err := stream.Send(&auctionPB.Update{NodeId: s.id, LeaderId: s.leaderId}); err != nil {
		log.Fatalf("Error sending join statement")
	}

	// Receive updates from leader:
	go s.receive(stream)
}

func (s *Server) UpdateNodes(nodeStream auctionPB.Nodes_UpdateNodesServer) error {
	replJoin, err := nodeStream.Recv()
	if err == io.EOF {
		log.Printf("This is an EOF Error")
		return err
	}
	if err != nil {
		log.Printf("This is another error")
		return err
	}
	log.Printf("Replica Node id #%d is ready for updates", replJoin.NodeId)

	// Add replica to map
	s.nodes[replJoin.NodeId] = nodeStream
	// Enqueue to leader queue
	s.leaderQueue = append(s.leaderQueue, replJoin.NodeId)
	s.broadcastUpdate()
	defer s.removeReplica(replJoin.NodeId)

	for {
		ack, err := nodeStream.Recv()
		if err == io.EOF {
			log.Printf("This is an EOF error receiving stream from replica node")
			//return err
			break
		}
		if err != nil {
			log.Printf("Error receiving stream from replica node")
			break
		}
		log.Printf("Received acknowledge from node #%d on version #%d", ack.NodeId, ack.Version)
	}
	return nil
}

func (s *Server) removeReplica(nodeId int32) {
	delete(s.nodes, nodeId)
	for i, v := range s.leaderQueue {
		if v == nodeId {
			s.leaderQueue = remove(s.leaderQueue, i)
		}
	}
	s.broadcastUpdate()
	log.Printf("Replica #%d is dead..", nodeId)
}

// Helper method to remove from slice
func remove(slice []int32, s int) []int32 {
	return append(slice[:s], slice[s+1:]...)
}

func (s *Server) Bid(ctx context.Context, bidReq *auctionPB.BidRequest) (*auctionPB.BidReply, error) {
	s.muHighBid.Lock()
	defer s.muHighBid.Unlock()
	s.muHighBidder.Lock()
	defer s.muHighBidder.Unlock()
	log.Printf("Received bid from client id #%d on amount: %d", bidReq.ClientId, bidReq.Amount)
	if !s.auctionLive {
		comment := "The auction is closed.."
		rep := &auctionPB.BidReply{Outcome: auctionPB.Outcome_FAIL, Comment: &comment}
		// Kan nil have problemer med næstløbende swtich statements?
		return rep, nil
	}

	var comment string
	var outcome auctionPB.Outcome

	// check bid with highest bid:
	hiBid := s.highBid
	switch {
	case hiBid < bidReq.Amount:
		comment = fmt.Sprintf("Your bid on amount: %d is accepted", bidReq.Amount)
		outcome = auctionPB.Outcome_SUCCESS
		s.highBid = bidReq.Amount
		s.highBidder = bidReq.ClientId
		s.version++
		s.broadcastUpdate()
	case hiBid == bidReq.Amount:
		comment = fmt.Sprintf("Your bid is equal to highest bid, but you were too slow..")
		outcome = auctionPB.Outcome_FAIL
	case hiBid > bidReq.Amount:
		comment = fmt.Sprintf("Your bid on amount: %d was too low", bidReq.Amount)
		outcome = auctionPB.Outcome_FAIL
	default:
		comment = fmt.Sprintf("You have f...ed something up")
		outcome = auctionPB.Outcome_EXCEPTION
	}

	reply := &auctionPB.BidReply{Outcome: outcome, Comment: &comment}
	return reply, nil
}

func (s *Server) Result(ctx context.Context, resReq *auctionPB.ResultRequest) (*auctionPB.ResultReply, error) {
	s.muHighBid.Lock()
	defer s.muHighBid.Unlock()
	s.muHighBidder.Lock()
	defer s.muHighBidder.Unlock()
	var comment string
	if !s.auctionLive {
		if s.highBid == 0 {
			comment = "The auction is closed and no bids recorded"
		} else {
			comment = fmt.Sprintf("The auction is closed. Winner is id #%d", s.highBidder)
		}
	} else {
		comment = fmt.Sprintf("The auction is live. Current highest bidder is id #%d", s.highBidder)
	}

	reply := &auctionPB.ResultReply{Comment: comment, HighestBid: s.highBid}
	return reply, nil
}

func (s *Server) broadcastUpdate() {
	log.Printf("Broadcasting update for version #%d to replica nodes...", s.version)
	for id, node := range s.nodes {
		if err := node.Send(&auctionPB.Update{Version: s.version,
			LeaderId:      s.leaderId,
			HighestBid:    s.highBid,
			AuctionLive:   s.auctionLive,
			Crashes:       s.crashes,
			HighestBidder: s.highBidder,
			Nodes:         s.leaderQueue}); err != nil {
			log.Printf("Error broadcasting to node #%d.. Is it dead?", id)
		}
	}
}

func (s *Server) receive(stream auctionPB.Nodes_UpdateNodesClient) {
	for {
		update, err := stream.Recv()
		// Connection to leader is dead:
		if err != nil {
			if s.leaderQueue[0] == s.id {
				// Become leader
				s.leaderId = s.id
				// remove yourself from leader queue
				s.leaderQueue = remove(s.leaderQueue, 0)
				log.Printf("The leader is dead.. Node #%d is now the new leader", s.id)
				s.launchServer()
				return
			} else {
				log.Printf("The leader is dead.. I am not worthy (yet).\nAttempting to dial new leader..")
				// TODO Maybe decrease this? or make it a for loop of e.g. 5 attempts
				time.Sleep(time.Second * 5)
				connectionToLeader(s)
				return
			}
		}
		s.leaderId = update.LeaderId
		s.auctionLive = update.AuctionLive
		s.highBid = update.HighestBid
		s.crashes = update.Crashes
		s.version = update.Version
		s.highBidder = update.HighestBidder
		s.leaderQueue = update.Nodes
		log.Printf("Got update from leader. Now on version %d", s.version)
		// Acknowledge leader with updated information
		if err := stream.Send(&auctionPB.Update{
			Version:       s.version,
			LeaderId:      s.leaderId,
			AuctionLive:   s.auctionLive,
			HighestBid:    s.highBid,
			HighestBidder: s.highBidder,
			NodeId:        s.id,
			Nodes:         s.leaderQueue}); err != nil {
			log.Fatalf("Something went wrong sending acknowledge to leader ")
		}
	}
}

// TODO: Update port, since it is irellevant when replicas acts as clients
// TODO: tidy up the multiple port listening
func (s *Server) launchServer() {

	log.Printf("Attemps to create listener on port 5400")
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:5400"))
	if err != nil {
		log.Fatalf("Failed to listen on port %d", s.port)
		return
	}

	list, err := net.Listen("tcp", fmt.Sprintf("localhost:5000"))
	if err != nil {
		log.Fatalf("Failed to listen on port 5000")
		return
	}

	// grpc server options:
	var opts []grpc.ServerOption

	// spin grpc server:
	grpcNodesServer := grpc.NewServer(opts...)
	grpcAuctionServer := grpc.NewServer(opts...)
	//auctionPB.RegisterAuctionServer(grpcNodesServer, s)
	auctionPB.RegisterNodesServer(grpcNodesServer, s)
	auctionPB.RegisterAuctionServer(grpcAuctionServer, s)
	go func() {
		// Serve incoming requests:
		if err := grpcNodesServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	go func() {
		// Serve incoming requests:
		if err := grpcAuctionServer.Serve(list); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	log.Printf("Server is listening on port %d", s.port)
	s.parseInput()
}

func (s *Server) connectToLeader() auctionPB.NodesClient {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	// maybe timeout is useful:
	//timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()

	log.Printf("Trying to dial 5400")
	conn, err := grpc.Dial(fmt.Sprintf(":5400"), opts...)
	if err != nil {
		log.Fatalf("Failed to dial on port: 5400")
	}
	leader := auctionPB.NewNodesClient(conn)
	log.Printf("Successfully connected to the leader")
	return leader
}

// setLog sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate(*serverName+*port+"Log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile(*serverName+*port+"log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}

func (s *Server) parseInput() {
	fmt.Println("Press <start> or <end> to start/end auction")
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Shit happenend reading input")
		}
		input = strings.TrimSpace(input)
		switch input {
		case "start":
			if s.auctionLive {
				log.Printf("Auction is already live\n")
				continue
			} else {
				s.highBidder = 0
				s.highBid = 0
				s.version = 0
				s.auctionLive = true
				s.broadcastUpdate()
			}
		case "end":
			if !s.auctionLive {
				log.Printf("Auction is already finished")
			} else {
				s.auctionLive = false
				s.broadcastUpdate()
			}
		}
	}
}
