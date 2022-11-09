System Requirements:

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