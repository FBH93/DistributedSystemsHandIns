package main

import (
	"fmt"
	"time"
)

var chClientToServer = make(chan packet, 200)
var chServerToClient = make(chan int, 200)

var serverListening = make(chan bool)
var establishHandshake = make(chan bool)
var respondHandshake = make(chan bool)

type packet struct {
	content     rune
	sequenceNum int
	lastPacket  bool
}

func client(message string) {
	var serverAvailable = <-serverListening

	if serverAvailable {
		fmt.Printf("Contacting server...\n")
		fmt.Printf("Sending SYN packet to server...\n")
		establishHandshake <- true
		var handshakeEstablished = <-respondHandshake
		if handshakeEstablished {
			fmt.Printf("Received ACK acknowledgement from server. \n")
			fmt.Printf("Handshake established. Ready to send packets\n")
			chars := []rune(message) //split the message into slice of characters(runes)
			for i := 0; i < len(chars); i++ {
				var p packet
				if i < len(chars)-1 {
					p = packet{chars[i], i, false}
				} else { // if packet is the last in message, flip lastPacket to true
					p = packet{chars[i], i, true}
				}
				chClientToServer <- p
				<-chServerToClient //Wait until server confirms receival of packet. This is where we can handle package loss
			}

		}
	}
}

func server() {
	serverListening <- true //Server is listening
	fmt.Printf("Server listening.... \n")
	for {
		<-establishHandshake
		fmt.Printf("Server received SYN packet by client\n")
		fmt.Printf("Responding to client with SYN-ACK packet\n")
		respondHandshake <- true
		fmt.Printf("Server acknowledged handshake. Ready to receive packets\n")
		receivedMessage := ""
		var receivedChars []rune
		for {
			receivedPacket := <-chClientToServer
			chServerToClient <- receivedPacket.sequenceNum
			fmt.Printf("Server received packet with content '%c' and seq '%d'\n", receivedPacket.content, receivedPacket.sequenceNum)
			receivedChars = append(receivedChars[:receivedPacket.sequenceNum], receivedPacket.content) //insert the received rune, at the position of the received sequenceNum in the Rune[] in order to handle message reordering
			if receivedPacket.lastPacket {
				fmt.Printf("All packets received")
				break
			}
		}
		receivedMessage = string(receivedChars)
		fmt.Printf("Server received following message: %s \n", receivedMessage)
		fmt.Printf("Server listening again.... \n")
		serverListening <- true //Server returns to listening

	}

	//recreate message
}

func main() {

	go server()
	go client("Hello World")
	go client("This is another message")

	time.Sleep(5000 * time.Millisecond)
	fmt.Printf("Finish Program\n")
}
