# Documentation for ChittyChat

This is HandIn 3 submission for group frjokr, consisting of

* Jonas Brock-Hanash
* Frederik Bøye Henriksen
* Kristian Moltke Reitzel

To run the solution for the ChittyChat system use the command

`go run server/server.go -ServerName <enter a server name here> -port <enter a port here>`

`go run client/client.go -name <enter a client name here> -port <enter a port here>`

Alternatively, build the solution for easy run.

`go build client/client.go` then run the .exe file with flags in command line `client.exe -name <name> -port <port>`
`go build server/server.go` then run the .exe file with flags in command line `server.exe -ServerName <name> -port <port>`

in separate command prompts in the HandIn3-folder. Multiple clients may be started.

When running the files, the optional flags `-name` and `-port` may be left out to revert to default values. If the port on the server is changed, the clients must use the same port.

## System Architecture

We use bidirectional streaming. The stream opens when the client connects to the server, and stays open for the duration of the chat,so that we know a client has disconnected from the server when the stream closes.

We use a server-client architecture. All our client nodes communicate with our server node, who then handles the parsing of messages to other client nodes.

## Implemented RPC methods

We have a single rpc method, that allows for a client to open a stream between the client and server, that allows for chat messages to be sent and received.

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
[!Sequence Diagram](ChittyChat/Assets/Sequence Diagram.svg)

## Git Repository

The source code for HandIn3 can be found at <https://github.com/FBH93/DistributedSystemsHandIns>

## System Logs

System logs can be found in the subfolder Assets, but for convenience we have compieled an image with the logs.

[!System logs](ChittyChat/Assets/Chatlogs.jpg)

The interesting scenario is as follows:
Start server
Client 1 joins
Client 2 joins
Client 1 sends message
Client 3 joins
Client 2 sends a message
Client 2 leaves
Client 3 sends message
Client 1 leaves
Client 3 sends a message
Client 3 leaves
Server closure

When running the ChittyChat program, new log files will be dynamically created, in the folder that the call to client.go/exe and server.go/exe is made from. If the log file already exists, it will be overwritten.

## System and Technical requirements

[Documented here](Requirements.md)