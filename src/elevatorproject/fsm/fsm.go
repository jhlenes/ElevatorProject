package main

import (
	"fmt"
	"time"
	"os"
	"os/signal"

	def "elevatorproject/definitions"
	"elevatorproject/driver/elevio"
	"elevatorproject/ordermanager"
	"elevatorproject/scheduler"
)

var Elevator struct {
	Floor     int
	Dir       def.Direction
	Behaviour ElevatorBehaviour
	ID        string
}

type ElevatorBehaviour int

// Elevator behaviours
const (
	Idle     ElevatorBehaviour = 0
	DoorOpen                   = 1
	Moving                     = 2
)

var doorTimerResetCh chan bool = make(chan bool)

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
		case a := <-drvButtons:
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

		case floor := <-drvFloors:
			fmt.Printf("%+v\n", floor)
			onFloorArrival(floor)

		case a := <-drvObstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(def.Stop)
			} else {
				elevio.SetMotorDirection(def.Up)
			}

		case a := <-drvStop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < def.NumFloors; f++ {
				for b := def.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
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
}


func onFloorArrival(newFloor int) {
	Elevator.Floor = newFloor
	elevio.SetFloorIndicator(newFloor)
	if scheduler.ShouldStop(Elevator.Floor, Elevator.Dir) {
		elevio.SetMotorDirection(def.Stop)
		elevio.SetDoorOpenLamp(true)
		scheduler.ClearOrders(Elevator.Floor, Elevator.Dir)
		doorTimerResetCh <- true
		Elevator.Behaviour = DoorOpen
		setAllLights()
	}
}

func onDoorTimeout() {
	Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(Elevator.Dir)
	if Elevator.Dir == def.Stop {
		Elevator.Behaviour = Idle
	} else {
		Elevator.Behaviour = Moving
		elevio.SetMotorDirection(Elevator.Dir)
	}
}

func setAllLights() {
	for floor := 0; floor < def.NumFloors; floor++ {
		for btn := def.ButtonType(0); btn < def.NumButtons; btn++ {
			elevio.SetButtonLamp(btn, floor, ordermanager.GetOrder(floor, btn))
		}
	}
}

func doorTimer(resetCh chan bool) {
	timer := time.NewTimer(def.DoorTimeout * time.Millisecond)
	timer.Stop()
	for {
		select {
		case <-resetCh:
			timer.Reset(def.DoorTimeout * time.Millisecond)
		case <-timer.C:
			onDoorTimeout()
		}
	}
}

// safeShutdown shutdowns the program in a safe way when terminataed by user (ctrl+c)
func safeShutdown() {
	var c = make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
	elevio.SetMotorDirection(def.Stop)
	fmt.Println("User terminated the program")
	os.Exit(1)
}