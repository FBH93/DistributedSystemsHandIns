package main

import (
	"flag"
	"fmt"
	pb "github.com/FBH93/DistributedSystemsHandIns/HandIn3/ChittyChat"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"time"
)

type Server struct {
	pb.UnimplementedChittyChatServer
	name    string
	port    string
	clients map[string]bool // Set of clients
	//TODO: Add mutex
}

// Flags:
var serverName = flag.String("name", "default server", "Server Name")
var port = flag.String("port", "5400", "Server Port")

func main() {
	flag.Parse()

	fmt.Println("--- Server is starting ---")

	go launchServer()

	for {
		time.Sleep(time.Second * 5)
	}
}

func launchServer() {
	log.Printf("Server %s: Attempts to create listener on port %s\n", *serverName, *port)

	// Create listener lis tcp on given port or default port 5400
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *serverName, *port, err)
		return
	}

	// Optional options for grpc server:
	var opts []grpc.ServerOption

	// Create pb server (not yet ready to accept requests yet)
	grpcServer := grpc.NewServer(opts...)

	// Make a server instance using the name and port from the flags
	server := &Server{
		name:    *serverName,
		port:    *port,
		clients: make(map[string]bool),
	}

	pb.RegisterChittyChatServer(grpcServer, server)

	log.Printf("Server %s: Listening on port %s\n", *serverName, *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

// TODO: Implement error if client name already exists
func (s *Server) addClient(clientName string) {
	s.clients[clientName] = true
}

func (s *Server) removeClient(clientName string) {
	delete(s.clients, clientName)
}

func (s *Server) Chat(server pb.ChittyChat_ChatServer) error {
	clientReq, err := server.Recv()
	if err == io.EOF {
		return err
	}
	if err != nil {
		return err
	}

	cliName := clientReq.ClientName
	// Add client:
	s.addClient(cliName)

	log.Printf(cliName + "Has joined the ChittyChat")
	if err := server.Send(&pb.ChatResponse{Msg: cliName + "has joined the ChittyChat"}); err != nil {
		log.Printf("Broadcast err: %v", err)
	}

	defer s.removeClient(cliName)

	for {
		response, err := server.Recv()
		if err != nil {
			log.Printf("recv err: %v", err)
			break
		}
		log.Printf("Broadcast: %s", response.Msg)
		if err := server.Send(&pb.ChatResponse{Msg: response.Msg}); err != nil {
			log.Printf("Broadcast err: %v", err)
		}
	}
	return nil
}
