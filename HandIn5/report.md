# Introduction

The goal of this hand-in is to implement, using Go, a simple auction system that is resilient to at most one node failure. 

# Architecture

The architecture is set up in such a way that the client need not know anything about the state of the system of servers. It is the responsibility of the servers to always provide the service at a particular port. If a leader dies, another replica should take over and start listening on the same port as the previous leader - and have a state consistent with the previous leader. This means that the distributed nature of the system is transparent to clients.

## Client

For this reason the code for the client (`Client.go`) is quite straightforward. It dials the service at port 5000 and then allows input of bids using stdin, which are subsequently sent to the server running on that particular port. Again, the client does not know nor care about replicas or which exact server instance is handling individual request.

## Server

The server (`Server.go`) has all logic for handling incoming bids from clients, keeping states in sync across replicas and handling failure of replicas though election of a new leader. A server acting as leader is to always a have a client-facing service running on port 5000 and a server-facing service on port 5400, that other replicas will dial. 

The leader maintains 

* a map `nodes` of all the replicas that are connected to it
* Highest bid amount `highBid` and the identifier of the highest bidder `highBidder`

Whenever there is a change to the state of the server, it will broadcast its updated state to all other replicas (for instance in the event of a new and higher bid, or in the event of a new replica joining the system). Each replica subsequently updates its own state to match that of the one broadcasted by the leader.

### Node failures



# Running the system

 Start servers by running

​	`go run server/server.go -port 5400` (initially acting as leader)

​	`go run server/server.go -port 5401` (acting as replica)

​	and so on, in order of sequentially increasing port numbers...

Start any number of clients using 

​	`go run client/client.go -id 1`

​	`go run client/client.go -id 2`

​	and so on...

In the client terminal, you can make a bet by typing a number. Inputting any non-integer value will send a result request to the server. 

# Correctness

## 1: Consistency

The system satisfies linearizability. 

Before the server sends acknowledgement for a bid back to a client, it requires the bid to be reflected in the state of all other replicas. This way, we always provide the most recently written value for subsequent reads. 

## 2: Protocol correctness

The system correctly handles crashes of any client or server. 

Take the example of the leader failing. Backup replicas will know this by the closure of the stream between the leader and each of the backup replicas. The state (including the list of replicas) of the server is kept in sync across replicas during the lifespan of the leader. Thanks to this, we have consistent list of IDs of the potential new leaders amongst the remaining replicas. From this list, the nodes agree that the server with the lowest ID becomes the new leader. This replica thus becomes a leader, starts serving clients on port 5000 and the remaining replicas on 5400, and normal operation of the system is resumed.

In `log.txt`, we demonstrate the an instance of the system running. 