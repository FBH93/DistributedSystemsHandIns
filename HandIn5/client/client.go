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
	logfile := setLog() //print log to a log.txt file instead of the console
	defer logfile.Close()

	parseId, _ := strconv.ParseInt(*clientId, 10, 32)
	id := int32(parseId)

	c := &Client{
		id: id,
	}
	log.Printf("Client %d attempting to join auction server", c.id)
	c.connectToServer()
	c.parseInput()

}

func (c *Client) parseInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Shit happenend reading input")
		}
		input = strings.TrimSpace(input)
		parseInt, _ := strconv.ParseInt(input, 10, 32)
		amount := int32(parseInt)

		c.bid(amount)
	}
}

func (c *Client) connectToServer() {
	// Dial options:
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	// Time out on the connection
	timeContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Dial the server to get a connection:
	log.Printf("Attempts to dial auction server")
	conn, err := grpc.DialContext(timeContext, fmt.Sprintf(":5000"), opts...)
	if err != nil {
		log.Printf("Failed to dial: %v\n", err)
		return
	}

	c.server = auctionPB.NewAuctionClient(conn)
	log.Printf("The connection is: %s\n", conn.GetState().String())
}

func (c *Client) bid(amount int32) {
	bid := &auctionPB.BidRequest{
		Amount:   amount,
		ClientId: c.id,
	}

	ack, err := c.server.Bid(context.Background(), bid)
	if err != nil {
		log.Printf("Something went wrong: %v", err)
	}

	log.Printf("Got ack from server:\nComment: %s\nOutcome: %v", *ack.Comment, ack.Outcome)
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
