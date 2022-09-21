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
		establishHandshake <- true
		var handshakeEstablished = <-respondHandshake
		if handshakeEstablished {
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
				<-chServerToClient //Wait until server confirms receival of packet
			}

		}
	}
}

func server() {
	serverListening <- true //Server is listening
	fmt.Printf("Server listening.... \n")
	for {
		<-establishHandshake
		fmt.Printf("Server received request for handshake\n")
		respondHandshake <- true
		fmt.Printf("Server acknowledged handshake. Ready to receive packets\n")
		receivedMessage := ""
		var receivedChars []rune
		for {
			receivedPacket := <-chClientToServer
			chServerToClient <- receivedPacket.sequenceNum
			fmt.Printf("Server received packet with content '%c' and seq '%d'\n", receivedPacket.content, receivedPacket.sequenceNum)
			receivedChars = append(receivedChars, receivedPacket.content)
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

	go client("Hello World")
	go server()
	go client("This is another message")

	//Two processes communicate with each other

	//take the message and break it into small pieces and send

	//The middleware should simulate the network. It can delay or lose packets.

	time.Sleep(5000 * time.Millisecond)
	fmt.Printf("Finish Program\n")
}
