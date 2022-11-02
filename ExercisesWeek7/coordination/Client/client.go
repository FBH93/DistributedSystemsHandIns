package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	pb "github.com/FBH93/DistributedSystemsHandIns/ExercisesWeek7/Client"
	"google.golang.org/grpc"
)

type Client struct {
	pb.UnimplementedCoordinationClient
	name          string
	port          string
	clients       map[string]pb.client // Set of clients and streams
	lampTime      int32
	mutexLampTime sync.Mutex
	mutexClient   sync.Mutex
}

// Flags:
var clientID = flag.String("id", "noID", "ID")
var port = flag.String("port", "5400", "Server Port")

func main() {
	flag.Parse() //Parse the flags from command line to server.

	//COMMENT OUT THESE TWO LINES TO REMOVE LOGGING TO TXT
	logfile := setLog()   //print log to a log.txt file instead of the console
	defer logfile.Close() //Close the log file when server closes.

	fmt.Println("--- Client is starting ---")
	go launchClient()

	for {
		time.Sleep(time.Second * 5)
	}
}

// launchServer sets up the grpc server with serverName to listen on the specified port.
func launchClient() {
	fmt.Printf("INFO: Server %s: Attempts to create listener on port %s\n", *clientID, *port)

	// Create listener lis tcp on given port or default port 5400
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", *port))
	if err != nil {
		log.Printf("Server %s: Failed to listen on port %s: %v", *clientID, *port, err)
		return
	}

	// Optional options for grpc server:
	var opts []grpc.ServerOption
	// Create pb server (not yet ready to accept requests yet)
	grpcClient := pb.NewCoordinationClient(opts...)

	// Make a server instance using the name and port from the flags
	client := &Client{
		name:     *clientID,
		port:     *port,
		clients:  make(map[string]Client),
		lampTime: 0,
	}

	pb.RegisterCoordinationServer(grpcClient, client)

	log.Printf("[T:%d] Client %s: Listening on port %s\n", client.lampTime, *clientID, *port)
	if err := grpcClient.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

// addClient adds client to clients with cliStream as its value
// Updates lampTime.
/*
func (s *Server) addClient(clientName string, cliStream pb.ChittyChat_ChatServer, cliTime int32) {
	s.mutexClient.Lock()
	s.clients[clientName] = cliStream
	s.mutexClient.Unlock()
	s.increaseLamptime(cliTime)
}


// removeClient removes client from clients together with its stream.
// Updates lampTime.
func (s *Server) removeClient(clientName string, cliTime int32) {
	s.mutexClient.Lock()
	delete(s.clients, clientName)
	s.mutexClient.Unlock()
	s.increaseLamptime(cliTime)
	leaveMsg := fmt.Sprintf("Participant %s left the server at server Lamport time %d \n", clientName, s.lampTime)
	log.Printf("[T:%d] %s", s.lampTime, leaveMsg) //log that client has left server.
	s.broadcast(leaveMsg)                         //broadcast to all connected clients, that a client has left.
}
*/

// increaseLamptime evaluates and updates lampTime.
func (s *Client) increaseLamptime(receivedTime int32) {
	s.mutexLampTime.Lock()
	defer s.mutexLampTime.Unlock()
	fmt.Printf("DEBUG: Evaluating client time %d vs received time %d \n", s.lampTime, receivedTime)
	if s.lampTime > receivedTime {
		s.lampTime++
		fmt.Printf("DEBUG: Increased lamptime by 1 to %d", s.lampTime)
	} else {
		s.lampTime = receivedTime + 1
		fmt.Printf("DEBUG: Increased lamptime to %d based on received time %d + 1 \n", s.lampTime, receivedTime)
	}
}

// broadcast a message to all connected clients
func (s *Client) broadcast(msg string) {
	s.increaseLamptime(s.lampTime)                         //Increase time before broadcasting
	log.Printf("[T:%d] Broadcasting: %s", s.lampTime, msg) //Log the message that is about to be broadcast
	for _, client := range s.clients {
		if err := client.Send(&pb.ChatResponse{Msg: msg, Time: s.lampTime}); err != nil {
			log.Printf("Broadcast error: %v", err)
		}
	}
}

// Chat is the main chat method.
// Adds a client, removes a client when connection is lost,
// receives chat messages and everything to all connected clients.
/*
func (s *Server) Chat(cliStream pb.ChittyChat_ChatServer) error {
	clientReq, err := cliStream.Recv()
	if err == io.EOF {
		return err
	}
	if err != nil {
		return err
	}

	cliName := clientReq.ClientName
	cliTime := clientReq.Time

	//JOIN CLIENT
	s.addClient(cliName, cliStream, cliTime)
	log.Printf("[T:%d] "+cliName+" Has joined the ChittyChat", s.lampTime) //log which client has joined, at some time.
	joinMsg := fmt.Sprintf("Participant %s joined Chitty-Chat at server Lamport time %d", cliName, s.lampTime)
	s.broadcast(joinMsg)

	//When client connection is lost, remove client
	defer s.removeClient(cliName, cliTime)

	//RECEIVE MESSAGE
	for {
		response, err := cliStream.Recv()
		if err != nil {
			fmt.Printf("ERROR: recv err: a stream closed \n")
			break
		}
		s.increaseLamptime(response.Time) //Increase cliStream time, based on time received from client.
		log.Printf("[T:%d] Server received message %s from %s", s.lampTime, response.Msg, response.ClientName)

		//Broadcast received message to clients.
		s.broadcast(response.Msg) //Broadcast msg received, to all clients.
	}

	//If the chat() function gives an error, return nil, you f'ed up.
	return nil
}
*/
// setLog sets the logger to use a log.txt file instead of the console.
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("Serverlog.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	file, err := os.OpenFile("ServerLog.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(file)
	return file
}
