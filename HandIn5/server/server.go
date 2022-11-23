package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
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
	version     int32
	nodes       map[int32]auctionPB.Nodes_UpdateNodesServer // map over nodes and streams
	leaderQueue []int32                                     // Queue of potential leaders
	auctionLive bool
	highBidder  int32
	highBid     int32
	lock        sync.Mutex
}

func main() {
	flag.Parse()

	parsePort, _ := strconv.ParseInt(*port, 10, 32)
	ownPort := int32(parsePort)
	s := &Server{
		name:        *serverName,
		port:        ownPort,
		id:          ownPort - 5400,
		nodes:       make(map[int32]auctionPB.Nodes_UpdateNodesServer),
		highBid:     0,
		leaderId:    0,
		leaderQueue: []int32{},
	}

	//COMMENT OUT THESE TWO LINES TO REMOVE LOGGING TO TXT
	//logfile := setLog() //print log to a log.txt file instead of the console
	//defer logfile.Close()

	// Make server listening on port 5400 the first leader when starting program
	if s.port == 5400 {
		s.auctionLive = false
		s.launchServer()
		// Otherwise, connect to leader:
	} else {
		connectToLeader(s)
	}

	// Keep server alive:
	for {
		time.Sleep(time.Second * 5)
	}
}

// connectToLeader establishes connection to the leader and receives
func connectToLeader(s *Server) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	log.Printf("Node #%d: Trying to dial 5400", s.id)
	conn, err := grpc.Dial(fmt.Sprintf(":5400"), opts...)
	if err != nil {
		log.Fatalf("Node #%d: Failed to dial on port: 5400", s.id)
	}
	leader := auctionPB.NewNodesClient(conn)
	log.Printf("Node #%d: Successfully connected to the leader", s.id)

	// Receive stream from leader:
	stream, err := leader.UpdateNodes(context.Background())
	if err != nil {
		log.Printf("Node #%d: Error getting stream from leader", s.id)
	}

	// Send join statement to leader:
	if err := stream.Send(&auctionPB.Update{NodeId: s.id, LeaderId: s.leaderId}); err != nil {
		log.Fatalf("Node #%d: Error sending join statement", s.id)
	}

	// Receive updates from leader:
	go s.receive(stream)
}

// UpdateNodes is used by the leader to receive streams and ack's from replicas
func (s *Server) UpdateNodes(nodeStream auctionPB.Nodes_UpdateNodesServer) error {
	// Get replica
	replJoin, err := nodeStream.Recv()
	if err != nil {
		log.Printf("Node #%d: Error receiving stream from replica node", s.id)
		return err
	}
	log.Printf("Node #%d: Replica Node id #%d has joined and ready for updates", s.id, replJoin.NodeId)
	// Add replica to map
	s.nodes[replJoin.NodeId] = nodeStream
	// Enqueue to leader queue
	s.leaderQueue = append(s.leaderQueue, replJoin.NodeId)
	// broadcast new replica to the system:
	s.broadcastUpdate()
	// remove replica when stream terminates
	defer s.removeReplica(replJoin.NodeId)

	// Receive acks from replicas:
	for {
		ack, err := nodeStream.Recv()
		if err != nil {
			log.Printf("Node #%d: Replica #%d is dead..", s.id, replJoin.NodeId)
			break
		}
		log.Printf("Node #%d: Received acknowledge from node #%d on version #%d", s.id, ack.NodeId, ack.Version)
	}
	return nil
}

// removeReplica removes replica and broadcasts update to other replicas
func (s *Server) removeReplica(nodeId int32) {
	delete(s.nodes, nodeId)
	for i, v := range s.leaderQueue {
		if v == nodeId {
			s.leaderQueue = remove(s.leaderQueue, i)
		}
	}
	s.broadcastUpdate()
}

// Helper method to remove from slice and preserve order
func remove(slice []int32, s int) []int32 {
	return append(slice[:s], slice[s+1:]...)
}

// Bid handles bids from auction clients
// Returns an outcome and comment to the client
func (s *Server) Bid(ctx context.Context, bidReq *auctionPB.BidRequest) (*auctionPB.BidReply, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	log.Printf("Node #%d: Received bid from client id #%d on amount: %d", s.id, bidReq.ClientId, bidReq.Amount)
	// Auction not live:
	if !s.auctionLive {
		comment := "The auction is closed.."
		rep := &auctionPB.BidReply{Outcome: auctionPB.Outcome_FAIL, Comment: &comment}
		// Kan nil have problemer med næstløbende swtich statements?
		return rep, nil
	}

	var comment string
	var outcome auctionPB.Outcome

	// check client bid with highest bid:
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

	// reply outcome and comment to client:
	reply := &auctionPB.BidReply{Outcome: outcome, Comment: &comment}
	return reply, nil
}

// Result returns the state of the auction, highest bid and the highest bidder
// return 0 for highest bid and bidder if no bid has been placed
func (s *Server) Result(ctx context.Context, resReq *auctionPB.ResultRequest) (*auctionPB.ResultReply, error) {
	log.Printf("Node #%d: Received result request", s.id)
	s.lock.Lock()
	defer s.lock.Unlock()
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

// broadcastUpdate broadcasts the current state of the leader to all replicas
func (s *Server) broadcastUpdate() {
	log.Printf("Node #%d: Broadcasting update for version #%d to replica nodes...", s.id, s.version)
	for id, node := range s.nodes {
		if err := node.Send(&auctionPB.Update{Version: s.version,
			LeaderId:      s.leaderId,
			HighestBid:    s.highBid,
			AuctionLive:   s.auctionLive,
			HighestBidder: s.highBidder,
			Nodes:         s.leaderQueue}); err != nil {
			log.Printf("Node #%d: Error broadcasting to node #%d.. Is it dead?", s.id, id)
		}
	}
}

// receive receives updates from the leader
func (s *Server) receive(stream auctionPB.Nodes_UpdateNodesClient) {
	for {
		update, err := stream.Recv()
		// Connection to leader is dead:
		if err != nil {
			// Your turn to become leader. Launch new servers
			if s.leaderQueue[0] == s.id {
				// Become leader
				s.leaderId = s.id
				// remove yourself from leader queue
				s.leaderQueue = remove(s.leaderQueue, 0)
				log.Printf("Node #%d: The leader is dead.. Node #%d is now the new leader", s.id, s.id)
				s.launchServer()
				return
				// Not your turn to become leader, wait and dial the new leader
			} else {
				log.Printf("Node #%d: The leader is dead.. I am not worthy (yet).\nAttempting to dial new leader..", s.id)
				time.Sleep(time.Second * 2)
				connectToLeader(s)
				return
			}
		}
		// Set your state to the leader's
		s.leaderId = update.LeaderId
		s.auctionLive = update.AuctionLive
		s.highBid = update.HighestBid
		s.version = update.Version
		s.highBidder = update.HighestBidder
		s.leaderQueue = update.Nodes

		log.Printf("Node #%d: Got update from leader. Now on version %d", s.id, s.version)

		// Acknowledge leader with updated information
		if err := stream.Send(&auctionPB.Update{
			Version:       s.version,
			LeaderId:      s.leaderId,
			AuctionLive:   s.auctionLive,
			HighestBid:    s.highBid,
			HighestBidder: s.highBidder,
			NodeId:        s.id,
			Nodes:         s.leaderQueue}); err != nil {
			log.Fatalf("Node #%d: Something went wrong sending acknowledge to leader", s.id)
		}
	}
}

// launchServer starts a server for replicas on port 5400 and a server for clients bidding on port 5000
// and creates a reader for starting/stopping an auction
func (s *Server) launchServer() {

	log.Printf("Node #%d: Attemps to create listener on port 5400", s.id)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:5400"))
	if err != nil {
		log.Fatalf("Node #%d: Failed to listen on port %d", s.id, s.port)
		return
	}

	list, err := net.Listen("tcp", fmt.Sprintf("localhost:5000"))
	if err != nil {
		log.Fatalf("Node #%d: Failed to listen on port 5000 and 5400", s.id)
		return
	}

	// grpc server options:
	var opts []grpc.ServerOption

	// spin grpc servers
	grpcNodesServer := grpc.NewServer(opts...)
	grpcAuctionServer := grpc.NewServer(opts...)
	auctionPB.RegisterNodesServer(grpcNodesServer, s)
	auctionPB.RegisterAuctionServer(grpcAuctionServer, s)
	// Serve incoming replica requests:
	go func() {
		if err := grpcNodesServer.Serve(lis); err != nil {
			log.Fatalf("Node #%d: Failed to serve: %v", s.id, err)
		}
	}()

	// Serve incoming auction requests:
	go func() {
		if err := grpcAuctionServer.Serve(list); err != nil {
			log.Fatalf("Node #%d: Failed to serve: %v", s.id, err)
		}
	}()
	log.Printf("Node #%d: Server is listening on port 5000 for auction bids", s.id)
	log.Printf("Node #%d: Server is listening on port 5400 for replica ack's", s.id)

	// Parse input on server-side to start/stop auction:
	s.parseInput()
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

// parseInput allows server to start or end an auction from cmd-line
func (s *Server) parseInput() {
	fmt.Println("Press <start> or <end> to start/end auction")
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Node #%d: Shit happenend reading input", s.id)
		}
		input = strings.TrimSpace(input)
		switch input {
		case "start":
			if s.auctionLive {
				log.Printf("Node #%d: Auction is already live\n", s.id)
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
				log.Printf("Node #%d: Auction is already finished", s.id)
			} else {
				s.auctionLive = false
				s.broadcastUpdate()
			}
		}
	}
}
