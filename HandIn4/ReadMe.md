# Hand In 4

## Instructions

Run the peers by opening three separate command prompts in the `HandIn4` folder and running each of these in one of them:

`go run . 0` for client 0 on port 5000

`go run . 1` for client 1 on port 5001

`go run . 2` for client 2 on port 5002

When the client is running, entering anything in a peer's command line (even a blank line) will cause the peer to try entering the critical section. 

Note how the artificial delay of 4 seconds in each execution of the critical makes it such that you can try to obtain permission while another peer is in the midst of the critical section execution.

### Handling entering the critical section

A peer may ping other peers freely at any time. A ping is a request to enter the critical section, and when a response has been received from all other peers, it may enter the critical section. 

## Discussion of implementation & algorithm

The implementation in our solution is loosely based on Ricart & Agrawala's algorithm - however, without the parts of the algorithm that guarantees ordering, as this is not the focus of this hand in. 

The state of a particular peer is stored in its corresponding struct, with the bool variables `wanted` and `held` indicating the state. Initially, both are false for all peers.  

When a peer $p1$ wishes to enter a critical section, its `wanted` is changed to true and a request is sent to all other peers. Once $N-1$ nodes have responded positively, $p1$ may proceed into the critical section. 

So long as either `wanted` or `held` for $p1$ is true, $p1$ will not respond positively to any incoming requests to enter the critical section, thus blocking other nodes from accessing the critical section currently.  

## Establishing contact between peers

The system is hardcoded to consist of three peers (however, this number trivial to change in line 53 of `main.go`). 

Each peer runs on a separate port. For this reason, we can conveniently use the port as the unique ID for each peer in our `clients` map - a part of the peer struct. 

A peer's own port is determined by the number provided as an argument `arg1` to the go program by $\text{arg1}+5000$, giving us peers running on successive ports $5000, 5001, 5002, ...$

Each peer dials all other peers upon start up - excluding itself - starting from port $5000$. By providing `grpc.WithBlock()` in the `grpc.Dial()` call, we make the dialing attempts blocking, thus giving us time to boot up the other peers.

## System Requirements:

> R1: Implement a system with a set of peer nodes, and a Critical Section, that represents a sensitive system operation. Any node can at any time decide it wants access to the Critical Section. Critical section in this exercise is emulated, for example by a print statement, or writing to a shared file on the network.

A node may enter the critical section at any time, by entering anything in the command prompt. The log will print updates for the peer's progress in entering the critical section.

> R2: Safety: Only one node at the same time is allowed to enter the Critical Section

When a peer wants to enter the critical section, it will first set its own state to `wanted = true`, to represent that it wants the acquire the lock. While the state is `wanted = true` any incoming requests received by the peer will be put in a waiting position (busy wait), i.e. not responded to immediately. 

After setting the state `wanted = true`, the peer will then try pinging all other connected clients. 
If positive responses have been received from all, the state will change to `wanted = false` and `held = true` to represent that the peer is now holding the lock to the critical section (and thus no longer only wanting the lock).
If a single negative response is received (for example due to time-out), the peer will retry entering the critical section after a short delay.
The peer will then do the critical section (consisting, in this case, of a sleep and a print statement to represent the time taken to perform some action, such as a database query), and once complete it will return to state `held = false`, thus letting the peer respond to all pings that it received while it was trying to enter the critical section.

This ensures only 1 peer may enter the critical section at the same time.

>  R2: Liveliness: Every node that requests access to the Critical Section, will get access to the Critical Section (at some point in time)

By nature of the loop in `ping()` method, we ensure that the peer $p1$ will eventually respond to all peers that pinged it - both in the event of a timeout or having successfully completed the critical section.
If $p1$ is in the critical section, it will eventually finish, and respond to other peers. If $p1$ does unexpectedly not respond within 10 loops (≈10 seconds), we can assume $p1$ is in a deadlock waiting for another peer, and forces $p1$ to give up the attempt to enter the critical section and retry after a short wait, thus giving other peers a chance to enter the critical section in the meantime.

This way we ensure that all who request to enter the critical section will *eventually* succeed.

# Example

This section will demonstrate the functioning of the program by providing the logs of an instance.   