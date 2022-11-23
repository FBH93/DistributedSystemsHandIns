package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	auction "github.com/FBH93/DistributedSystemsHandIns/HandIn5/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
)

// run using go run Client/Client.go -id=?
var clientId = flag.Int("id", 0, "Client id")
var serverPort = flag.String("server", "5000", "Auction server")

var server auction.AuctionClient //the server
var ServerConn *grpc.ClientConn  //the server connection

func main() {
	//parse flag/arguments
	flag.Parse()
	fmt.Printf("Starting client with id: %v \n", *clientId)

	fmt.Println("------ CLIENT APP ------")

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
	log.Printf("client %v: Attempts to dial on port %s\n", *clientId, *serverPort)
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
		Id:  int32(*clientId),
	}
	var ctx = context.Background()
	ack, err := server.Bid(ctx, request)
	if err != nil {
		// log.Println(err)
		log.Printf("Client %v: no response from the server, attempting to reconnect to next port and resend message...", *clientId)
		ctx.Done() // "cancel" existing bid request
		serverPortInt, _ := strconv.ParseInt(*serverPort, 10, 32)
		*serverPort = strconv.FormatInt(1+serverPortInt, 10)
		ConnectToServer()
		server.Bid(context.Background(), request)

	} else {
		if ack.Ack == true {
			fmt.Printf("Success, acknowledgement for bid was returned by server \n")
		}
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
