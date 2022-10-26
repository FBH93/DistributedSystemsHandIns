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

### change this after lamport

```golang
service ChittyChat {
  rpc Chat (stream ChatRequest) returns (stream ChatResponse) {};
}

message ChatRequest {
  string msg = 1;
  string clientName = 2;
}

message ChatResponse {
  string msg = 1;
}
```

## Lamport Timestamps

Describe how you have implemented the calculation of the Lamport timestamps

## Diagram

Provide a diagram, that traces a sequence of RPC calls together with the Lamport timestamps, that corresponds to a chosen sequence of interactions: Client X joins, Client X Publishes, ..., Client X leaves. Include documentation (system logs) in your appendix.

## Git Repository

The source code for HandIn3 can be found at <https://github.com/FBH93/DistributedSystemsHandIns>

## System Logs

Include system logs, that document the requirements are met, in the appendix of your report
