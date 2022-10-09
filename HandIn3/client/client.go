package main

import (
	"flag"
	pb "github.com/FBH93/DistributedSystemsHandIns/HandIn3/ChittyChat"
	"google.golang.org/grpc"
)

// Flags:
var clientName = flag.String("name", "Default Client", "Name of client")
var serverPort = flag.String("port", "5400", "tcp server")

var server pb.ChittyChatClient
var serverConn *grpc.ClientConn

func main() {
	flag.Parse()

}
