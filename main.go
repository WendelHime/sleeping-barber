package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type BarberStatus int

const (
	AwakeBarberStatus    BarberStatus = iota
	SleepingBarberStatus BarberStatus = iota
)

type Barber struct {
	Ready  *sync.Mutex
	Status BarberStatus
}

type Customer struct {
	HairCutDuration time.Duration
}

type BarberShop struct {
	Barber            Barber
	MaxWaitingRoom    int
	WaitingRoomChairs chan (Customer)
	WaitingRoomMutex  *sync.Mutex
	WaitGroup         *sync.WaitGroup
	ClosedMutex       *sync.Mutex
	Closed            bool
}

func NewBarberShop(waitingRoomChairs int) *BarberShop {
	barber := Barber{
		Ready:  new(sync.Mutex),
		Status: AwakeBarberStatus,
	}

	return &BarberShop{
		Barber:            barber,
		MaxWaitingRoom:    waitingRoomChairs,
		WaitingRoomChairs: make(chan Customer, waitingRoomChairs),
		WaitingRoomMutex:  new(sync.Mutex),
		WaitGroup:         new(sync.WaitGroup),
		ClosedMutex:       new(sync.Mutex),
		Closed:            false,
	}
}

func BarberFlow(barberShop *BarberShop) {
	var customer Customer
	for {
		// check if there's a customer waiting
		barberShop.WaitingRoomMutex.Lock()
		barberShop.Barber.Ready.Lock()
		fmt.Println("Barber checks the waiting room")
		customersWaiting := len(barberShop.WaitingRoomChairs)
		// if there's no customers, go sleep
		if customersWaiting == 0 {
			barberShop.ClosedMutex.Lock()
			if barberShop.Closed {
				fmt.Println("BarberShop closed! Time to go home")
				barberShop.ClosedMutex.Unlock()
				barberShop.Barber.Ready.Unlock()
				barberShop.WaitingRoomMutex.Unlock()
				break
			}
			barberShop.ClosedMutex.Unlock()
			fmt.Println("no customers, time to sleep!")
			barberShop.Barber.Status = SleepingBarberStatus
		}
		barberShop.Barber.Ready.Unlock()
		barberShop.WaitingRoomMutex.Unlock()

		// if there's a customer get the first, update status and cut hair
		customer = <-barberShop.WaitingRoomChairs
		// awake barber and accomodate customer
		barberShop.Barber.Ready.Lock()
		if barberShop.Barber.Status == SleepingBarberStatus {
			fmt.Println("Client awake barber")
			barberShop.Barber.Status = AwakeBarberStatus
		}
		barberShop.Barber.Ready.Unlock()

		// cut hair
		fmt.Println("Barber start to cut hair")
		time.Sleep(customer.HairCutDuration)
		fmt.Println("Hair cut finished!")
	}
	barberShop.WaitGroup.Done()
}

func randomCustomerAppears() Customer {
	randTimeBeforeArrives := rand.Intn(30)
	time.Sleep(time.Duration(randTimeBeforeArrives) * time.Millisecond)
	randHairCutDuration := rand.Intn(20)
	return Customer{HairCutDuration: time.Duration(randHairCutDuration) * time.Millisecond}
}

func CustomerFlow(barberShop *BarberShop) {
	for {
		customer := randomCustomerAppears()
		barberShop.WaitingRoomMutex.Lock()
		barberShop.ClosedMutex.Lock()
		fmt.Println("Customer checks waiting room and barber status")
		chairsAvailable := barberShop.MaxWaitingRoom - len(barberShop.WaitingRoomChairs)
		if chairsAvailable > 0 && !barberShop.Closed {
			fmt.Println("customer join the waiting room")
			barberShop.WaitingRoomChairs <- customer
		} else if barberShop.Closed {
			fmt.Println("customer leaving because the barbre shop is closed")
		} else {
			fmt.Println("customer leaving because there are no seats available :(")
		}
		if barberShop.Closed {
			barberShop.ClosedMutex.Unlock()
			barberShop.WaitingRoomMutex.Unlock()
			break
		}
		barberShop.ClosedMutex.Unlock()
		barberShop.WaitingRoomMutex.Unlock()
	}
	barberShop.WaitGroup.Done()
}

func main() {
	var numberOfFreeWRSeats int
	var openTime int

	flag.IntVar(&numberOfFreeWRSeats, "numberOfFreeWRSeats", 1, "number of free seats on the barber shop")
	flag.IntVar(&openTime, "openTime", 800, "barber open time")
	flag.Parse()

	barberShop := NewBarberShop(numberOfFreeWRSeats)

	rand.Seed(time.Now().UnixNano())

	go func() {
		time.Sleep(time.Duration(openTime) * time.Millisecond)
		barberShop.ClosedMutex.Lock()
		barberShop.Closed = true
		barberShop.ClosedMutex.Unlock()
		close(barberShop.WaitingRoomChairs)
	}()
	go BarberFlow(barberShop)
	go CustomerFlow(barberShop)
	barberShop.WaitGroup.Add(2)
	barberShop.WaitGroup.Wait()
}
