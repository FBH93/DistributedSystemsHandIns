package main

import (
	"fmt"
	"time"
)

var ch1 = make(chan int)
var ch2 = make(chan int)
var ch3 = make(chan int)
var ch4 = make(chan int)
var ch5 = make(chan int)

var maxEats = 3 //Number of times each philospher has to eat.

func philosopher(ID int, fork1 chan int, fork2 chan int) {
	ate := 1
	thoughts := 0
	fmt.Printf("Philosopher %d is thinking... \n", ID)
	for ate <= maxEats {
		var y = <-fork1 //Receive fork1 when it is available (i.e something in the channel)
		var x = <-fork2 //receive fork2 when it is available

		fmt.Printf("Philosopher %d is eating with fork %d and %d for the %d time \n", ID, y, x, ate) //prints when Philosopher eats successfully
		ate++
		thoughts++
		fmt.Printf("Philosopher %d is thinking again with thought %d... \n", ID, thoughts) //Prints when Philosopher is done eating
		fork1 <- y                                                                         //Put fork1 down
		fork2 <- x                                                                         //put fork2 down

	}
}

func fork(ID int, ch chan int) {
	ch <- ID //Fork broadcasts it is available
}

func main() {
	fmt.Println("Starting Program...")
	go philosopher(1, ch5, ch1)
	go philosopher(2, ch1, ch2)
	go philosopher(3, ch2, ch3)
	go philosopher(4, ch3, ch4)
	go philosopher(5, ch4, ch5)
	go fork(1, ch1)
	go fork(2, ch2)
	go fork(3, ch3)
	go fork(4, ch4)
	go fork(5, ch5)

	time.Sleep(10000 * time.Millisecond)
	fmt.Printf("Finish program")
}
