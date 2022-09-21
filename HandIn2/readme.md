# a) What are packages in your implementation? What data structure do you use to transmit data and meta-data? Marshalling??
Packets are modeled using a struct `packet` that contain the content 

# b) Does your implementation use threads or processes? 
Threads
## Why is it not realistic to use threads?
Channels do not correctly model the real way of sending and receiving packets over the internet. With channels in Go we guarentee to receive packets in the same order we send them i.e. as a FIFO queue. In network communication, the packets are however not guarenteed
    
# c) How do you handle message re-ordering?
Instead of the server simply appending the received contents to the servers receievedMessage, we place it in an array/slice at the correct position (i.e. the index reported by the client in the packets' sequenceNum)

# d) How do you handle message loss?
We don't. Currently the implementation guarentees that it -only- prints if all packets are received, due to the server sending a confirmation upon receiving each packet. If no acknowledgement is received, the current implementation of the client will hang forever waiting for confirmation. 
If we were to handle message loss, it would involve having the client wait for acknowledgement from the server after each packet is sent. If a certain time has passed with no acknowledgement, the client should resend the packet and repeat the process.

# e) Why is the 3-way handshake important?
It ensures that both server and client is ready to send/receive data