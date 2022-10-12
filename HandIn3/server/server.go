package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	pb "github.com/FBH93/DistributedSystemsHandIns/HandIn3/ChittyChat"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedChittyChatServer
	name     string
	port     string
	clients  map[string]pb.ChittyChat_ChatServer // Set of clients
	lampTime int32
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
		name:     *serverName,
		port:     *port,
		clients:  make(map[string]pb.ChittyChat_ChatServer),
		lampTime: 0,
	}

	pb.RegisterChittyChatServer(grpcServer, server)

	log.Printf("Server %s: Listening on port %s\n", *serverName, *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

// TODO: Implement error if client name already exists
func (s *Server) addClient(clientName string, server pb.ChittyChat_ChatServer, cliTime int32) {
	s.clients[clientName] = server

	s.increaseLamptime(cliTime) //Increase lamptime
}

func (s *Server) removeClient(clientName string, cliTime int32) {
	delete(s.clients, clientName)
	s.increaseLamptime(cliTime) //increase lamptime
}

func (s *Server) increaseLamptime(receivedTime int32) {
	if s.lampTime > receivedTime {
		s.lampTime++
	} else {
		s.lampTime = receivedTime + 1
	}
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
	cliTime := clientReq.Time
	// Add client:
	s.addClient(cliName, server, cliTime) //add client and increase lampTime when client has been connected to server

	log.Printf("[T:%d] "+cliName+" Has joined the ChittyChat", s.lampTime) //log which client has joined, at some time.

	s.increaseLamptime(s.lampTime) //Increase time before broadcasting a client has joined.

	log.Printf("[T:%d] Broadcasting: %s has joined the chat \n", s.lampTime, cliName)
	for _, client := range s.clients {
		if err := client.Send(&pb.ChatResponse{Msg: cliName + " has joined the ChittyChat", Time: s.lampTime}); err != nil {
			log.Printf("Broadcast error: %v", err)
		}
	}

	//if err := server.Send(&pb.ChatResponse{Msg: cliName + "has joined the ChittyChat"}); err != nil {
	//	log.Printf("Broadcast err: %v", err)
	//}

	defer s.removeClient(cliName, cliTime) //remove client and increase time when removed.

	for {
		response, err := server.Recv()
		if err != nil {
			log.Printf("recv err: %v", err)
			break
		}
		s.increaseLamptime(response.Time) //Increase server time, based on time received from client.

		//Broadcast msg received, to all clients.
		s.increaseLamptime(s.lampTime) //before broadcast, increase server lamptime
		log.Printf("[T:%d] Broadcasting: %s \n", s.lampTime, response.Msg)
		for _, client := range s.clients {
			if err := client.Send(&pb.ChatResponse{Msg: response.Msg, Time: s.lampTime}); err != nil {
				log.Printf("Broadcast error: %v", err)
			}
		}
	}
	return nil
}

//func (s *Server) broadcast(request pb.ChatRequest) {
//
//}
