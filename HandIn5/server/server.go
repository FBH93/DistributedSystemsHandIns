package main

import (
	"context"
	"flag"
	"fmt"
	auctionPB "github.com/FBH93/DistributedSystemsHandIns/HandIn5/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
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
	auctionPB.UnimplementedNodesServer
	name     string
	id       int32
	port     int32
	leaderId int32
	leader   bool
	version  int32
	crashes  int32
	ctx      context.Context
	nodes    map[int32]auctionPB.Nodes_UpdateNodesServer // map over nodes and streams
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
		name:     *serverName,
		port:     ownPort,
		id:       ownPort - 5400,
		ctx:      ctx,
		nodes:    make(map[int32]auctionPB.Nodes_UpdateNodesServer),
		highBid:  0,
		leaderId: 0,
	}
	// make server listening on port 5400 the first leader when starting program
	if s.port == 5400 {
		s.leader = true
		// FOR TESTING. TODO: Implement timer on auction
		s.auctionLive = true
		s.launchServer()
	} else {
		s.leader = false
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

	// Keep server alive:
	for {
		time.Sleep(time.Second * 5)
	}
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

	for {
		ack, err := nodeStream.Recv()
		if err == io.EOF {
			log.Printf("This is an EOF error receiving stream from replica node")
			return err
		}
		if err != nil {
			log.Printf("This is another error receiving stream from replica node")
		}
		log.Printf("Received acknowledge from node #%d on version #%d", ack.NodeId, ack.Version)
	}
	return nil
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

	hiBid := s.highBid
	switch {
	case hiBid < bidReq.Amount:
		comment = fmt.Sprintf("Your bid on amount: %d is accepted", bidReq.Amount)
		outcome = auctionPB.Outcome_SUCCESS
		s.highBid = bidReq.Amount
		s.highBidder = bidReq.ClientId
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

	s.broadcastUpdate()
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
			comment = "The auction is closed and no auctions recorded"
		} else {
			comment = fmt.Sprintf("The auction is closed. Winner is id #%d", s.highBidder)
		}
	} else {
		comment = fmt.Sprintf("The auction is live. Highest bidder is id #%d", s.highBidder)
	}

	reply := &auctionPB.ResultReply{Comment: comment, HighestBid: s.highBid}
	return reply, nil
}

// TODO: NEXT STEP: Implement broadcast
func (s *Server) broadcastUpdate() {
	log.Printf("Broadcasting update for version #%d to replica nodes...", s.version)
	for id, node := range s.nodes {
		if err := node.Send(&auctionPB.Update{Version: s.version, LeaderId: s.leaderId, HighestBid: s.highBid, AuctionLive: s.auctionLive}); err != nil {
			log.Printf("Error broadcasting to node #%d.. Is it dead?", id)
		}
	}
}

// TODO Implement receive
func (s *Server) receive(stream auctionPB.Nodes_UpdateNodesClient) {
	for {
		update, err := stream.Recv()
		if err != nil {
			// Stream closed
			// TODO: Implement crashes
			if s.leaderId+1 == s.id {
				// Become leader
				s.leaderId = s.id
				s.crashes++
				log.Printf("The leader is dead.. Node #%d is now the new leader", s.id)
				s.launchServer()
			} else {
				log.Printf("The leader is dead.. I am not worthy (yet).\nAttempting to dial new leader..")
				// TODO Maybe decrease this? or make it a for loop of e.g. 5 attempts
				time.Sleep(time.Second * 5)
				// Exit out of this forever loop??
				// e.g. add return value, then return s.connectToLeader()
				s.connectToLeader()
			}
		}
		s.leaderId = update.LeaderId
		s.auctionLive = update.AuctionLive
		s.highBid = update.HighestBid
		// TODO: add crashes to proto
		s.version = update.Version
		log.Printf("Got update from leader. Now on version %d", s.version)
		// Acknowledge leader with updated information
		if err := stream.Send(&auctionPB.Update{Version: s.version, LeaderId: s.leaderId, AuctionLive: s.auctionLive, HighestBid: s.highBid, NodeId: s.id}); err != nil {
			log.Fatalf("Something went wrong sending acknowledge to leader ")
		}
	}
}

// TODO: Update port, since it is irellevant when replicas acts as clients
func (s *Server) launchServer() {

	log.Printf("Attemps to create listener on %d", s.port)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:5400"))
	if err != nil {
		log.Fatalf("Failed to listen on port %d", s.port)
		return
	}

	// grpc server options:
	var opts []grpc.ServerOption

	// spin grpc server:
	grpcServer := grpc.NewServer(opts...)
	//auctionPB.RegisterAuctionServer(grpcServer, s)
	auctionPB.RegisterNodesServer(grpcServer, s)
	log.Printf("NodesServer registrered")
	go func() {
		// Serve incoming requests:
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	log.Printf("Server is listening on port %d", s.port)
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
