package main

import (
	"bufio"
	"fmt"
	"net"
)

func check(err error, message string) {
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", message)
}

type ClientJob struct {
	name string
	conn net.Conn
}

func generateResponses(clientJobs chan ClientJob) {
	for {
		// Wait for the next job to come off the queue.
		clientJob := <-clientJobs

		// Do something thats keeps the CPU buys for a whole second.
		//for start := time.Now(); time.Now().Sub(start) < time.Second; {
		//}

		// Send back the response.
		clientJob.conn.Write([]byte("Hello, " + clientJob.name))
	}
}

func main() {
	clientJobs := make(chan ClientJob, 8)
	go generateResponses(clientJobs)

	ln, err := net.Listen("tcp", ":8080")
	check(err, "Server is ready.")

	for {
		conn, err := ln.Accept()
		check(err, "Accepted connection.")

		go func() {
			buf := bufio.NewReader(conn)

			for {
				name, err := buf.ReadString('\n')

				if err != nil {
					fmt.Printf("Client disconnected.\n")
					break
				}

				clientJobs <- ClientJob{name, conn}
			}
		}()
	}
}

/*
package main

import (
	"fmt"
	"time"
)

type Request interface{}

func handle(r Request) { fmt.Println(r.(int)) }

const RateLimitPeriod = time.Second
const RateLimit = 2 // most 200 requests in one minute

func handleRequests(requests <-chan Request) {
	quotas := make(chan time.Time, RateLimit)

	go func() {
		tick := time.NewTicker(RateLimitPeriod / RateLimit)
		defer tick.Stop()
		for t := range tick.C {
			select {
			case quotas <- t:
			default:
			}
		}
	}()

	for r := range requests {
		<-quotas
		go handle(r)
	}
}

func main() {
	requests := make(chan Request)
	go handleRequests(requests)
	// time.Sleep(time.Minute)
	for i := 0; ; i++ {
		requests <- i
	}
}
*/

/*
package main

import (
	"fmt"
	"time"
)

func Tick(d time.Duration) <-chan struct{} {
	// The capacity of c is best set as one.
	c := make(chan struct{}, 1)
	go func() {
		for {
			time.Sleep(d)
			select {
			case c <- struct{}{}:
			default:
			}
		}
	}()
	return c
}

func main() {
	t := time.Now()
	for range Tick(time.Second) {
		fmt.Println(time.Since(t))
	}
}
*/

/*
package main

import "fmt"

func main() {
	type Book struct{ id int }
	bookshelf := make(chan Book, 16)

	for i := 0; i < cap(bookshelf)*2; i++ {
		select {
		case bookshelf <- Book{id: i}:
			fmt.Println("succeeded to put book", i)
		default:
			fmt.Println("failed to put book +>", i)
		}
	}

	for i := 0; i < cap(bookshelf)*2; i++ {
		select {
		case book := <-bookshelf:
			fmt.Println("succeeded to get book", book.id)
		default:
			fmt.Println("failed to get book")
		}
	}
}
*/

/*
package main

import "runtime"

func DoSomething() {
	for {
		// do something ...

		runtime.Gosched() // avoid being greedy
	}
}

func main() {
	go DoSomething()
	go DoSomething()
	select {}
}
*/

/*
package main

import (
	"fmt"
	"os"
)

type Ball uint64

func Play(playerName string, table chan Ball) {
	var lastValue Ball = 1
	for {
		ball := <-table // get the ball
		fmt.Println(playerName, ball)
		ball += lastValue
		if ball < lastValue { // overflow
			os.Exit(0)
		}
		lastValue = ball
		table <- ball // bat back the ball
		//		time.Sleep(time.Second)
	}
}

func main() {
	table := make(chan Ball)
	go func() {
		table <- 1 // throw ball on table
	}()
	go Play("A:", table)
	Play("B:", table)
}
*/

/*
package main

import (
	"log"
	"math/rand"
	"time"
)

type Seat int
type Bar chan Seat

func (bar Bar) ServeCustomer(c int) {
	log.Print("customer#", c, " enters the bar")
	seat := <-bar // need a seat to drink
	log.Print("++ customer#", c, " drinks at seat#", seat)
	time.Sleep(time.Second * time.Duration(2+rand.Intn(6)))
	log.Print("-- customer#", c, " frees seat#", seat)
	bar <- seat // free seat and leave the bar
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// the bar has 10 seats.
	bar24x7 := make(Bar, 10)
	// Place seats in an bar.
	for seatId := 0; seatId < cap(bar24x7); seatId++ {
		// None of the sends will block.
		bar24x7 <- Seat(seatId)
	}

	for customerId := 2; ; customerId++ {
		time.Sleep(time.Second)
		go bar24x7.ServeCustomer(customerId)
	}

}
*/
/*
package main

import (
	"log"
	"time"
)

type T = struct{}

func worker(id int, ready <-chan T, done chan<- T) {
	<-ready // block here and wait a notification
	log.Print("Worker#", id, " starts.")
	// Simulate a workload.
	time.Sleep(time.Second * time.Duration(id+1))
	log.Print("Worker#", id, " job done.")
	// Notify the main goroutine (N-to-1),
	done <- T{}
}

func main() {
	log.SetFlags(0)

	ready, done := make(chan T), make(chan T)
	go worker(0, ready, done)
	go worker(1, ready, done)
	go worker(2, ready, done)

	// Simulate an initialization phase.
	time.Sleep(time.Second * 3 / 2)
	// 1-to-N notifications.
	ready <- T{}
	ready <- T{}
	ready <- T{}
	// Being N-to-1 notified.
	<-done
	<-done
	<-done
}

*/

/*
package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"sort"
)

func main() {
	values := make([]byte, 32*1024*1024)
	if _, err := rand.Read(values); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	done := make(chan struct{}, 10000) // can be buffered or not

	// The sorting goroutine
	go func() {
		sort.Slice(values, func(i, j int) bool {
			return values[i] < values[j]
		})
		// Notify sorting is done.
		done <- struct{}{}
	}()

	// do some other things ...

	<-done // waiting here for notification
	fmt.Println(values[255], values[len(values)-1])
}
*/

/*
package main

import (
	"fmt"
	"time"
)

func main() {
	n := 3

	// This is where we "make" the channel, which can be used
	// to move the `int` datatype
	out := make(chan int)

	// We still run this function as a goroutine, but this time,
	// the channel that we made is also provided
	go multiplyByTwo(n, out)

	// Once any output is received on this channel, print it to the console and proceed
	fmt.Println(<-out)
}

// This function now accepts a channel as its second argument...
func multiplyByTwo(num int, out chan<- int) {
	result := num * 2

	time.Sleep(time.Second * 2)
	//... and pipes the result into it
	out <- result
}
*/
