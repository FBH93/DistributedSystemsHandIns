# Hand In 4


## System Functionality

run the peers by opening cmd in the HandIn4 folder and using the command:

`go run . 0` for client 0 on port 5000
`go run . 1` for client 1 on port 5001
`go run . 2` for client 2 on port 5002

When the client is running, entering anything in a peer's command line (even a blank line) will cause the peer to try entering the critical section.

### Handling entering the critical section

A peer may ping other peers freely at any time. A ping is a request to enter the critical section, and when a response has been received from all other peers, it may enter the critical section. 

## System Requirements:

### R1: Implement a system with a set of peer nodes, and a Critical Section, that represents a sensitive system operation. Any node can at any time decide it wants access to the Critical Section. Critical section in this exercise is emulated, for example by a print statement, or writing to a shared file on the network.

A node may enter the critical section at any time, by entering anything in the cmd log. The log will print updates for the peer's progress in entering the critical section.

### R2: Safety: Only one node at the same time is allowed to enter the Critical Section 

When a peer wants to enter the critical section, it will first set its own state to `Wanted = true`, to represent it wants the acquire the lock.
While the state is `wanted = true` any pings received by the peer, will be put in a waiting position (busy wait), i.e. not responded to immedietly. 
After setting the state `wanted = true`, the peer will then ping all other connected clients. 
    If positive responses have been received from all, the state will change to `wanted = false` and `held = true` to represent that the peer is now holding the lock to the critical section (and thus no longer wanting the lock).
    If a single negative response is received (for example due to time-out), the peer will retry entering the critical section after a short delay.
The peer will then do the critical section (A sleep and a print statement to represent the time taken to perform some action, such as a database query), and once complete it will return to state `held = false`, thus letting the peer respond to all pings that it received while it was trying to enter the critical section.

This ensures only 1 peer may enter the critical section at the same time.

### R2: Liveliness: Every node that requests access to the Critical Section, will get access to the Critical Section (at some point in time)

By nature of the loop in ping() method, we ensure that the peer (p1) will eventually respond to all other peers that pinged it. (Either through timeout or having completed the critical section.)
if p1 is in the critical section, it will eventually finish, and respond to other peers. If p1 does not respond within 10 loops (10 seconds), we can assume p1 is in a deadlock waiting for another peer, and forces p1 to leave the critical section and retry after a short wait, thus giving other peers a chance to enter the critical section in the meantime.

This way we ensure that eventually all who request to enter the critical section will succeed, but they may have to wait a while.

## Technical Requirements:

Use Golang to implement the service's nodes
Provide a README.md, that explains how to start your system
Use gRPC for message passing between nodes
Your nodes need to find each other.  For service discovery, you can choose one of the following options
    Supply a file with IP addresses/ports of other nodes
    Enter IP address/ports through the command line -- We use this
    use a package for service discovery, like the Serf package 
Demonstrate that the system can be started with at least 3 nodes
Demonstrate using your system's logs,  a sequence of messages in the system, that leads to a node getting access to the Critical Section. You should provide a discussion of your algorithm, using examples from your logs.

Hand-in requirements:

Hand in a single report in a pdf file
Provide a link to a Git repo with your source code in the report
Include system logs, that document the requirements are met, in the appendix of your report