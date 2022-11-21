package main

import (
	"flag"
	auctionPB "github.com/FBH93/DistributedSystemsHandIns/HandIn5/grpc"
)

// Flags:
var clientId = flag.String("ID", "1", "Id of client")

type Client struct {
	id     int32
	server auctionPB.AuctionClient
}

func main() {

}
