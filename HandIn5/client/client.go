package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	auctionPB "github.com/FBH93/DistributedSystemsHandIns/HandIn5/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Flags:
var clientId = flag.String("id", "1", "Id of client")

type Client struct {
	id     int32
	server auctionPB.AuctionClient
}

func main() {
	flag.Parse()

	//COMMENT OUT THESE TWO LINES TO REMOVE LOGGING TO TXT
	//logfile := setLog() //print log to a log.txt file instead of the console
	//defer logfile.Close()
	parseId, _ := strconv.ParseInt(*clientId, 10, 32)
	id := int32(parseId)

	c := &Client{
		id: id,
	}
	log.Printf("Client #%d: Attempting to join auction server", c.id)
	c.connectToServer()
	c.parseInput()

}

// parseInput parses input from client
// any intput parsed as an int is a bid
// else it is a result request
func (c *Client) parseInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Client #%d: Shit happenend reading input", c.id)
		}
		input = strings.TrimSpace(input)
		parseInt, err := strconv.ParseInt(input, 10, 32)
		if err != nil {
			c.result()
		} else {
			amount := int32(parseInt)
			c.bid(amount)
		}
	}
}

// Connect to auction server
func (c *Client) connectToServer() {
	// Dial options:
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	// Time out on the connection
	timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Dial the server to get a connection:
	log.Printf("Client #%d: Attempts to dial auction server", c.id)
	conn, err := grpc.DialContext(timeContext, fmt.Sprintf(":5000"), opts...)
	if err != nil {
		log.Printf("Client #%d: Failed to dial: %v\n", c.id, err)
		return
	}

	c.server = auctionPB.NewAuctionClient(conn)
	log.Printf("Client #%d: The connection is: %s\n", c.id, conn.GetState().String())
}

// bid places a bid on the auction and receives an ack from server
func (c *Client) bid(amount int32) {
	log.Printf("Client #%d: Requesting bid...", c.id)
	bid := &auctionPB.BidRequest{
		Amount:   amount,
		ClientId: c.id,
	}

	ack, err := c.server.Bid(context.Background(), bid)
	if err != nil {
		log.Printf("Client #%d: Something went wrong: %v", c.id, err)
		log.Printf("Client #%d: Something went wrong: %v", c.id, err)
	}

	log.Printf("Client #%d: Got ack from server:\nComment: %s\nOutcome: %v", c.id, *ack.Comment, ack.Outcome)
}

// result queries the auction server for the current state of the auction
func (c *Client) result() {
	log.Printf("Client #%d: Requesting result...", c.id)
	request := &auctionPB.ResultRequest{}
	ack, err := c.server.Result(context.Background(), request)
	if err != nil {
		log.Printf("Client #%d: Something went wrong: %v", c.id, err)

	}
	log.Printf("Client #%d: Got result from server:\nComment: %s\nOutcome: %v", c.id, ack.Comment, ack.HighestBid)
}

// setLog sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate(*clientId+"Log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile(*clientId+"log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}
