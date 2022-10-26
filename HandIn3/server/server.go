package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	pb "github.com/FBH93/DistributedSystemsHandIns/HandIn3/ChittyChat"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedChittyChatServer
	name          string
	port          string
	clients       map[string]pb.ChittyChat_ChatServer // Set of clients
	lampTime      int32
	mutexLampTime sync.Mutex
	mutexClient   sync.Mutex
}

// Flags:
var serverName = flag.String("name", "default server", "Server Name")
var port = flag.String("port", "5400", "Server Port")

func main() {
	logfile := setLog() //print log to a log.txt file instead of the console
	defer logfile.Close()
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

	log.Printf("[T:%d] Server %s: Listening on port %s\n", server.lampTime, *serverName, *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

// TODO: Implement error if client name already exists
func (s *Server) addClient(clientName string, server pb.ChittyChat_ChatServer, cliTime int32) {
	s.mutexClient.Lock()
	s.clients[clientName] = server
	s.mutexClient.Unlock()
	s.increaseLamptime(cliTime) //Increase lamptime
}

// Helper to remove clients
func (s *Server) removeClient(clientName string, cliTime int32) {
	s.mutexClient.Lock()
	delete(s.clients, clientName)
	s.mutexClient.Unlock()
	s.increaseLamptime(cliTime) //increase lamptime
	leaveMsg := fmt.Sprintf("Participant %s left the server at server Lamport time %d \n", clientName, s.lampTime)
	log.Printf("[T:%d] %s", s.lampTime, leaveMsg) //log that client has left server.
	s.broadcast(leaveMsg)                         //broadcast to all connected clients, that a client has left.
}

// Helper to evaluate lamport time
func (s *Server) increaseLamptime(receivedTime int32) {
	s.mutexLampTime.Lock()
	defer s.mutexLampTime.Unlock()
	if s.lampTime > receivedTime {
		s.lampTime++
		fmt.Printf("DEBUG: increased lamptime \n")
	} else {
		s.lampTime = receivedTime + 1
		fmt.Printf("DEBUG: Increased lamptime based on higher value received time \n")
	}
}

// Broadcast a message to all connected clients
func (s *Server) broadcast(msg string) {
	s.increaseLamptime(s.lampTime)                         //Increase time before broadcasting
	log.Printf("[T:%d] Broadcasting: %s", s.lampTime, msg) //Log the message that is about to be broadcast
	for _, client := range s.clients {
		if err := client.Send(&pb.ChatResponse{Msg: msg, Time: s.lampTime}); err != nil {
			log.Printf("Broadcast error: %v", err)
		}
	}
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}

// MAIN CHAT FUNCTION
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

	//JOIN CLIENT
	s.addClient(cliName, server, cliTime)                                  //add client and increase lampTime when client has been connected to server
	log.Printf("[T:%d] "+cliName+" Has joined the ChittyChat", s.lampTime) //log which client has joined, at some time.
	joinMsg := fmt.Sprintf("Participant %s joined Chitty-Chat at server Lamport time %d", cliName, s.lampTime)
	s.broadcast(joinMsg) //Broadcast joinmsg to all clients.

	//LEAVE CLIENT
	defer s.removeClient(cliName, cliTime) //remove client and increase time when removed.

	//RECEIVE MESSAGE
	for {
		response, err := server.Recv()
		if err != nil {
			log.Printf("recv err: stream closed")
			break
		}
		s.increaseLamptime(response.Time) //Increase server time, based on time received from client.
		log.Printf("[T:%d]Server received message %s from %s", s.lampTime, response.Msg, response.ClientName)

		//Broadcast received message to clients.
		//sendMsg := fmt.Sprintf("[T:%d] Broadcasting: %s \n", s.lampTime, response.Msg)
		s.broadcast(response.Msg) //Broadcast msg received, to all clients.
	}

	//If the chat() function gives an error, return nil, you f'ed up.
	return nil
}
