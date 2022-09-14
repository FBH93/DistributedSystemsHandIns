package main

import (
	"fmt"
	"time"
)

// Each channel corresponds to each fork
var ch1Available = make(chan int)
var ch1PutDown = make(chan int)
var ch2Available = make(chan int)
var ch2PutDown = make(chan int)
var ch3Available = make(chan int)
var ch3PutDown = make(chan int)
var ch4Available = make(chan int)
var ch4PutDown = make(chan int)
var ch5Available = make(chan int)
var ch5PutDown = make(chan int)

var maxEats = 3 //Number of times each philospher has to eat.

func philosopher(ID int, fork1Avail chan int, fork1PutDown chan int, fork2Avail chan int, fork2PutDown chan int) {
	ate := 1
	thoughts := 0
	fmt.Printf("Philosopher %d is thinking... \n", ID)
	//loop continues until philosopher eats 3 times
	for ate <= maxEats {
		time.Sleep(50 * time.Millisecond) //To prove that the code is running concurrently
		var y = <-fork1Avail              //Receive fork1 when it is available (i.e something in the channel)
		var x int                         //variable declaration for fork2
		select {
		case x = <-fork2Avail: //receive fork2 if it is available
			fmt.Printf("Philosopher %d is eating with fork %d and %d for the %d time \n", ID, y, x, ate) //prints when Philosopher eats successfully
			ate++
			thoughts++
			fork1PutDown <- y                                                                  //Put fork1 down
			fork2PutDown <- x                                                                  //put fork2 down
			fmt.Printf("Philosopher %d is thinking again with thought %d... \n", ID, thoughts) //Prints when Philosopher is done eating
		default:
			fork1PutDown <- y //Put fork1 down if there is no fork2 available to avoid deadlock
			//continue          //loop
		}

	}
	fmt.Printf("Philosopher %d is finished eating \n", ID)
}

func fork(ID int, chAvailable chan int, chPutDown chan int) {
	chAvailable <- ID //Fork broadcasts it is *initially* available
	for {             //Checks if the fork has been put down.
		<-chPutDown
		fmt.Printf("Fork %d has been put down \n", ID)
		chAvailable <- ID //If fork has been put down, make it available again
	}
}

func main() {
	fmt.Println("Starting Program...")
	go fork(1, ch1Available, ch1PutDown)
	go fork(2, ch2Available, ch2PutDown)
	go fork(3, ch3Available, ch3PutDown)
	go fork(4, ch4Available, ch4PutDown)
	go fork(5, ch5Available, ch5PutDown)
	go philosopher(2, ch2Available, ch2PutDown, ch1Available, ch1PutDown)
	go philosopher(1, ch5Available, ch5PutDown, ch1Available, ch1PutDown)
	go philosopher(3, ch2Available, ch2PutDown, ch3Available, ch3PutDown)
	go philosopher(4, ch3Available, ch3PutDown, ch4Available, ch4PutDown)
	go philosopher(5, ch4Available, ch4PutDown, ch5Available, ch5PutDown)

	time.Sleep(5000 * time.Millisecond)
	fmt.Printf("Finish program")
}
