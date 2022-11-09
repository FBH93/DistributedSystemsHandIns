# Hand In 4


## System Functionality
### Handling entering the critical section
A peer may ping other peers freely at any time. A ping is a request to enter the critical section, and when a response has been received from all other peers, it may enter the critical section. 

When a peer wants to enter the critical section, it will first set its own state to `Wanted = true`, to represent it wants the acquire the lock.
While the state is `wanted = true` any pings received by the peer, will be put in a waiting position (busy wait), i.e. not responded to immedietly. 
After setting the state `wanted = true`, the peer will then ping all other connected clients. Once responses have been received from all, the state will change to `wanted = false` and `held = true` to represent that the peer is now holding the lock to the critical section (and thus no longer wanting the lock). 
The peer will then do the critical section (A sleep and a print statement to represent the time taken to perform some action, such as a database query), and once complete it will return to state `held = false`, thus letting the peer respond to all pings that it received while it was trying to enter the critical section.


## System Requirements:

R1: Implement a system with a set of peer nodes, and a Critical Section, that represents a sensitive system operation. Any node can at any time decide it wants access to the Critical Section. Critical section in this exercise is emulated, for example by a print statement, or writing to a shared file on the network.



R2: Safety: Only one node at the same time is allowed to enter the Critical Section 

R2: Liveliness: Every node that requests access to the Critical Section, will get access to the Critical Section (at some point in time)

Technical Requirements:

Use Golang to implement the service's nodes
Provide a README.md, that explains how to start your system
Use gRPC for message passing between nodes
Your nodes need to find each other.  For service discovery, you can choose one of the following options
Supply a file with IP addresses/ports of other nodes
Enter IP address/ports through the command line
use a package for service discovery, like the Serf package 
Demonstrate that the system can be started with at least 3 nodes
Demonstrate using your system's logs,  a sequence of messages in the system, that leads to a node getting access to the Critical Section. You should provide a discussion of your algorithm, using examples from your logs.
Hand-in requirements:

Hand in a single report in a pdf file
Provide a link to a Git repo with your source code in the report
Include system logs, that document the requirements are met, in the appendix of your report