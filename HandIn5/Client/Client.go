package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	auction "github.com/FBH93/DistributedSystemsHandIns/HandIn5/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Same principle as in client. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var serverPort = flag.String("server", "5000", "Auction server")

var server auction.AuctionClient //the server
var ServerConn *grpc.ClientConn  //the server connection

func main() {
	//parse flag/arguments
	flag.Parse()

	fmt.Println("--- CLIENT APP ---")

	//log to file instead of console
	//f := setLog()
	//defer f.Close()

	//connect to server and close the connection when program closes
	fmt.Println("--- join Server ---")
	ConnectToServer()
	defer ServerConn.Close()

	//start the input
	scanner := bufio.NewScanner(os.Stdin)
	var val int32 = 1
	for scanner.Scan() {
		//Do something when pressing enter in console
		val = val + 1
		bid(val)
	}
}

// connect to server
func ConnectToServer() {

	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	//dial the server, with the flag "server", to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, *serverPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", *serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	// makes a client from the server connection and saves the connection
	// and prints rather or not the connection was is READY
	server = auction.NewAuctionClient(conn)
	ServerConn = conn
	log.Println("the connection is: ", conn.GetState().String())
}

func bid(inputBid int32) {
	request := &auction.Bid{
		Bid: inputBid,
	}
	ack, err := server.Bid(context.Background(), request)
	if err != nil {
		log.Printf("Client %s: no response from the server, attempting to reconnect", *clientsName)
		log.Println(err)
	}

	// check if the server has handled the request correctly
	if ack.Ack == true {
		fmt.Printf("Success, the bid was received")
	} else {
		// something could be added here to handle the error
		// but hopefully this will never be reached
		fmt.Println("Oh no something went wrong :(")
	}
}

// Function which returns a true boolean if the connection to the server is ready, and false if it's not.
func conReady(s auction.AuctionClient) bool {
	return ServerConn.GetState().String() == "READY"
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
