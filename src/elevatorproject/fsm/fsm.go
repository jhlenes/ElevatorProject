package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	def "elevatorproject/definitions"
	"elevatorproject/driver/elevio"
	"elevatorproject/ordermanager"
	"elevatorproject/scheduler"
)

type ElevatorBehaviour int

// Elevator behaviours
const (
	Idle ElevatorBehaviour = iota
	DoorOpen
	Moving
)

var Elevator struct {
	Floor     int
	Dir       def.Direction
	Behaviour ElevatorBehaviour
	ID        string
}

// Channels used to reset timers
var doorTimerResetCh chan bool = make(chan bool)
var watchdogTimerResetCh chan bool = make(chan bool)

func main() {

	// Initializations
	elevio.Init(def.Addr, def.NumFloors)
	initFsm()

	// Channels
	drvButtons := make(chan def.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)
	drvStop := make(chan bool)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)

	// Listen to channels
	for {
		select {
		case button := <-drvButtons:
			onNewOrder(button)

		case floor := <-drvFloors:
			onFloorArrival(floor)

		case a := <-drvObstr:
			log.Printf("Obstruction: %+v\n", a)
			if a {
				elevio.SetMotorDirection(def.Stop)
			} else {
				elevio.SetMotorDirection(def.Up)
			}

		case a := <-drvStop:
			log.Printf("Stop: %+v\n", a)
			/*for f := 0; f < def.NumFloors; f++ {
				for b := def.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}*/
		}
	}
}

func initFsm() {
	Elevator.Floor = -1
	Elevator.Dir = def.Stop
	Elevator.Behaviour = Idle
	Elevator.ID = "A"

	go safeShutdown()
	go doorTimer(doorTimerResetCh)
	go watchdogTimer(watchdogTimerResetCh)
}

func onNewOrder(button def.ButtonEvent) {
	log.Printf("New request: %+v", button)
	resetWatchdogTimer()

	switch Elevator.Behaviour {
	case DoorOpen:
		if Elevator.Floor == button.Floor {
			doorTimerResetCh <- true
		} else {
			ordermanager.AddOrder(button.Floor, button.Button)
		}
	case Moving:
		ordermanager.AddOrder(button.Floor, button.Button)
	case Idle:
		if Elevator.Floor == button.Floor {
			elevio.SetDoorOpenLamp(true)
			Elevator.Behaviour = DoorOpen
			doorTimerResetCh <- true
		} else {
			ordermanager.AddOrder(button.Floor, button.Button)
			Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
			elevio.SetMotorDirection(Elevator.Dir)
			Elevator.Behaviour = Moving
		}
	}
	setAllLights()
}

func onFloorArrival(newFloor int) {
	log.Printf("Arrived at floor %v.", newFloor)
	resetWatchdogTimer()

	if Elevator.Floor == -1 { // elevator was initialized between floors
		Elevator.Dir = def.Stop
	}

	elevio.SetMotorDirection(Elevator.Dir) // make sure elevator is going the way is says it is

	Elevator.Floor = newFloor
	elevio.SetFloorIndicator(newFloor)
	if scheduler.ShouldStop(Elevator.Floor, Elevator.Dir) {
		elevio.SetMotorDirection(def.Stop)
		elevio.SetDoorOpenLamp(true)
		scheduler.ClearOrders(Elevator.Floor, Elevator.Dir)
		resetDoorTimer()
		Elevator.Behaviour = DoorOpen
		setAllLights()
	}
}

func onDoorTimeout() {
	log.Printf("Door timedout.")
	resetWatchdogTimer()

	Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(Elevator.Dir)
	if Elevator.Dir == def.Stop {
		Elevator.Behaviour = Idle
	} else {
		Elevator.Behaviour = Moving
	}
}

func onWatchdogTimeout() {
	log.Printf("Watchdog timedout.")
	resetWatchdogTimer()

	switch Elevator.Behaviour {
	case Idle:
		Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
		elevio.SetMotorDirection(Elevator.Dir)
		if Elevator.Dir == def.Stop {
			Elevator.Behaviour = Idle
		} else {
			Elevator.Behaviour = Moving
		}
	case DoorOpen:
		// TODO: Figure out if this can happen and what to do
	case Moving: // Elevator is stuck
		// try to restart motor
		elevio.SetMotorDirection(Elevator.Dir)
	}
}

func setAllLights() {
	for floor := 0; floor < def.NumFloors; floor++ {
		for btn := def.ButtonType(0); btn < def.NumButtons; btn++ {
			elevio.SetButtonLamp(btn, floor, ordermanager.HasOrder(floor, btn))
		}
	}
}

func resetDoorTimer() {
	doorTimerResetCh <- true
}

func doorTimer(resetCh chan bool) {
	timer := time.NewTimer(def.DoorTimeout * time.Millisecond)
	timer.Stop()
	for {
		select {
		case <-resetCh:
			timer.Reset(def.DoorTimeout * time.Millisecond)
		case <-timer.C:
			go onDoorTimeout()
		}
	}
}

func resetWatchdogTimer() {
	watchdogTimerResetCh <- true
}

func watchdogTimer(resetCh chan bool) {
	timer := time.NewTimer(def.WatchdogTimeout * time.Millisecond)
	timer.Stop()
	for {
		select {
		case <-resetCh:
			timer.Reset(def.WatchdogTimeout * time.Millisecond)
		case <-timer.C:
			go onWatchdogTimeout()
		}
	}
}

// safeShutdown shutdowns the program in a safe way when terminataed by user (ctrl+c)
func safeShutdown() {
	var c = make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
	elevio.SetMotorDirection(def.Stop)
	log.Println("User terminated the program.")
	os.Exit(1)
}
