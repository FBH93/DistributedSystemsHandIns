package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	pb "github.com/FBH93/DistributedSystemsHandIns/HandIn3/ChittyChat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Flags:
var clientName = flag.String("name", "Default Client", "Name of client")
var serverPort = flag.String("port", "5400", "tcp server")

var lampTime int32 = 0 //client starts with lamport time 0

var server pb.ChittyChatClient
var serverConn *grpc.ClientConn

func main() {
	flag.Parse()

	//COMMENT OUT THESE TWO LINES TO REMOVE LOGGING TO TXT
	//logfile := setLog() //print log to a log.txt file instead of the console
	//defer logfile.Close()

	log.Printf("[T:%d] Attempting to connect to server", lampTime)
	connectToServer()

	// Get stream from server
	stream, err := server.Chat(context.Background())
	if err != nil {
		log.Println(err)
		log.Printf("Could not connect to server. Is it running?")
		return
	}
	go receive(stream)

	// Ensure first message to server initializes the welcome message
	if err := stream.Send(&pb.ChatRequest{Msg: "", ClientName: *clientName}); err != nil {
		log.Fatal(err)
	}

	parseInput(stream)
}

// increaseLamptime evaluates and updates lampTime.
func increaseLamptime(receivedTime int32) {
	fmt.Printf("DEBUG: Evaluating client time %d vs received time %d \n", lampTime, receivedTime)
	if lampTime > receivedTime {
		lampTime++
		fmt.Printf("DEBUG: Increased lamptime by 1 to %d", lampTime)
	} else {
		lampTime = receivedTime + 1
		fmt.Printf("DEBUG: Increased lamptime to %d based on received time %d + 1 \n", lampTime, receivedTime)
	}
}

// receive and log stream from server
func receive(stream pb.ChittyChat_ChatClient) {
	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Fatal(err)
		}
		increaseLamptime(resp.Time) //before printing received msg, increase time
		log.Printf("[T:%d] %s", lampTime, resp.Msg)
	}
}

// connectToServer with specified dial options.
// increase lampTime when connection is successful
func connectToServer() {
	// Dial options:
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	// Time out on the connection
	timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Dial the server to get a connection:
	fmt.Printf("INFO: %s: Attempts to dial on port %s\n", *clientName, *serverPort)
	conn, err := grpc.DialContext(timeContext, fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Failed to dial: %v\n", err)
		return
	}

	server = pb.NewChittyChatClient(conn)
	serverConn = conn
	increaseLamptime(lampTime) //Client has connected to server, increase time.
	log.Printf("[T:%d] The connection is: %s\n", lampTime, conn.GetState().String())
}

// parseInput from client and send to server stream
func parseInput(stream pb.ChittyChat_ChatClient) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("INFO: --- You can now post a message to ChittyChat ---")

	for {
		//fmt.Printf("-> ")

		// Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			increaseLamptime(lampTime)                                                //Increase time when connection to server is lost/disconnected.
			log.Fatalf("[T:%d] Connection to server interrupted. Goodbye!", lampTime) //Client goodbye message.
		}
		// Trim input
		input = strings.TrimSpace(input)

		//Test if server is ready to receive stream
		if serverConn.GetState().String() != "READY" {
			log.Fatalf("Client %s: Something was wrong with the connection to the server :(", *clientName)
		}

		prefix := *clientName + ": "
		message := prefix + input
		increaseLamptime(lampTime) //Before sending message to server, increase lamptime
		log.Printf("[T:%d] sent message '%s'", lampTime, message)
		if err := stream.Send(&pb.ChatRequest{Msg: message, ClientName: *clientName, Time: lampTime}); err != nil {
			log.Fatal(err)
		}
	}
}

// setLog sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate(*clientName+"Log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile(*clientName+"log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
