# Documentation for ChittyChat
This is HandIn 3 submission for group frjokr, consisting of

* Jonas Brock-Hanash
* Frederik Bøye Henriksen
* Kristian Moltke Reitzel

To run the solution for the ChittyChat system use the command

`go run server/server.go -ServerName <enter a server name here> -port <enter a port here>`

`go run client/client.go -name <enter a client name here> -port <enter a port here>`

Alternatively, build the solution:

`go build client/client.go` then run the .exe file with flags in command line `client.exe -name <name> -port <port>`
`go build server/server.go` then run the .exe file with flags in command line `server.exe -ServerName <name> -port <port>`

in separate command prompts in the HandIn3-folder. Multiple clients may be started.

When running the files, the optional flags `-name` and `-port` may be left out to revert to default values. If the port on the server is changed, the clients must use the same port.

> WARNING: We have not taken into account multiple clients with the same name. This may create issues with with the log files, as they will write to the same log file. The chat client will still work though.

## System Architecture

We use bidirectional streaming. The stream opens when the client connects to the server, and stays open for the duration of the chat,so that we know a client has disconnected from the server when the stream closes.

We use a server-client architecture. All our client nodes communicate with our server node, who then handles the parsing of messages to other client nodes.

## RPC method

We have a single RPC method, that allows for a client to open a stream between the client and server, that allows for chat messages to be sent and received.

```golang
service ChittyChat {
  rpc Chat (stream ChatRequest) returns (stream ChatResponse) {};
}

message ChatRequest {
  string msg = 1;
  string clientName = 2;
  int32 time = 3;
}

message ChatResponse {
  string msg = 1;
  int32 time = 2;
}
```

## Lamport Timestamps

Each node (that is, a client or server) maintains a `lampTime` field, which represents the internal lamport time of that node. All nodes start with a `lampTime` of zero.

The lamport time of the sending party is included in all RPC messages.

Both clients and servers have a helper method `increaseLamptime(receivedTime int32)`, which is responsible for determining what to increase the `lampTime` to, and takes an argument `receivedTime` which is compared to the internal lamport time. This function is called whenever an event which should increase the lamport time takes place.

To easily see when lamport times are evaluated, we have included debug messages in the console. These do not get included in the log files.

The function either

1. increments `lampTime` by one in the event that the internal `lampTime` is larger than the received time OR
2. sets `lampTime` to the received time plus one

The events we decided to constitute an increase in lamport time are the following:

* A client joining
* A client leaving
* A client sending a `ChatRequest` to the server
* The server receiving a `ChatRequest` from a client
* The server broadcasting a message (*one* broadcast event may send a `ChatResponse` to *many* clients)
* A client receiving a `ChatResponse` from the server.

## Diagram

Below is a diagram illustrating an example sequence of events in ChittyChat.

The local lamport time is shown in brackets next to discrete events, and the lamport time in a message is shown in brackets next to the message type.

<img src="ChittyChat/Assets/Sequence Diagram.svg">

## Git Repository

The source code for HandIn3 can be found at <https://github.com/FBH93/DistributedSystemsHandIns>

## System Logs

System logs can be found in the subfolder Assets, but for convenience we have compieled an image with the logs.

![System logs](ChittyChat/Assets/Chatlogs.jpg)

The interesting scenario is as follows:

* Start server
* Client 1 joins
* Client 2 joins
* Client 1 sends message
* Client 3 joins
* Client 2 sends a message
* Client 2 leaves
* Client 3 sends message
* Client 1 leaves
* Client 3 sends a message
* Client 3 leaves
* Server closure

When running the ChittyChat program, new log files will be dynamically created, in the folder that the call to client.go/exe and server.go/exe is made from. If the log file already exists, it will be overwritten.

# System requirements

## R1: Chitty-Chat is a distributed service, that enables its clients to chat. The service is using gRPC for communication. You have to design the API, including gRPC methods and data types.  Discuss, whether you are going to use server-side streaming, client-side streaming, or bidirectional streaming?

We use bidirectional streaming. The stream opens when the client connects to the server, and stays open for the duration of the chat,so that we know a client has disconnected from the server when the stream closes

## R2: Clients in Chitty-Chat can Publish a valid chat message at any time they wish.  A valid message is a string of UTF-8 encoded text with a maximum length of 128 characters. A client publishes a message by making a gRPC call to Chitty-Chat

The program can handle sending messages of any size (that we tested), at any time through gRPC calls.

## R3: The Chitty-Chat service has to broadcast every published message, together with the current Lamport timestamp, to all participants in the system, by using gRPC. It is an implementation decision left to the students, whether a Vector Clock or a Lamport timestamp is sent

The server broadcasts any message received to all connected clients, along with a synchronized lamport timestamp. 

## R4: When a client receives a broadcasted message, it has to write the message and the current Lamport timestamp to the log

The client prints the received message along with the synchronized lamport time. (If a server sends a message at lamport time X, the client will print at X+1 because we see the receive action as a separate event in terms of lamport time)

## R5: Chat clients can join at any time

Chat clients can join the server at any time.

## R6: A "Participant X  joined Chitty-Chat at Lamport time L" message is broadcast to all Participants when client X joins, including the new Participant

The server will announce and broadcast when a user has joined.

## R7: Chat clients can drop out at any time

Chat clients can drop out from the chat, by simply closing the cmd-client. 

The client leaving will also receie a good-bye message in the log to confirm they have disconnected (if closed with control+C).

## R8: A "Participant X left Chitty-Chat at Lamport time L" message is broadcast to all remaining Participants when Participant X leaves

The server will broadcast when a client has left the server. 

# Technical Requirements

## Use gRPC for all messages passing between nodes

The server uses gRPC for all message passing.

## Use Golang to implement the service and clients

All code is written in golang.

## Every client has to be deployed as a separate process

Each client joins as a separate process via a new command prompt.

## Log all service calls (Publish, Broadcast, ...) using the log package

Both client and Server makes a log with a helpermethod SetLog()

![SetLog() method](ChittyChat/Assets/SetLog().JPG)
![SetLogDetails](ChittyChat/Assets/SetLog()Details.JPG)

## Demonstrate that the system can be started with at least 3 client nodes

See Chatlog example
![3Clients](ChittyChat/Assets/Chatlogs.jpg)

## Demonstrate that a client node can join the system

See Chatlog example

## Demonstrate that a client node can leave the system

See Chatlog example
