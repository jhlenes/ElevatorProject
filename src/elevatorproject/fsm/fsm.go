package main

import (
	"fmt"
	"time"

	"elevatorproject/definitions"
	"elevatorproject/driver/elevio"
	"elevatorproject/ordermanager"
	"elevatorproject/scheduler"
)

var Elevator struct {
	Floor     int
	Dir       definitions.Direction
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

var doorTimerResetCh chan bool

func main() {

	// Initializations
	elevio.Init(definitions.Addr, definitions.NumFloors)
	initFsm()

	// Channels
	drvButtons := make(chan definitions.ButtonEvent)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)
	drvStop := make(chan bool)
	doorTimerResetCh = make(chan bool)

	go elevio.PollButtons(drvButtons)
	go elevio.PollFloorSensor(drvFloors)
	go elevio.PollObstructionSwitch(drvObstr)
	go elevio.PollStopButton(drvStop)
	go doorTimer(doorTimerResetCh)

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
				setMotorDirection(definitions.Stop)
			} else {
				setMotorDirection(definitions.Up)
			}

		case a := <-drvStop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < definitions.NumFloors; f++ {
				for b := definitions.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}

func initFsm() {
	Elevator.Floor = -1
	Elevator.Dir = definitions.Stop
	Elevator.Behaviour = Moving
	Elevator.ID = "A"
	setMotorDirection(definitions.Up)
}

func setMotorDirection(dir definitions.Direction) {
	elevio.SetMotorDirection(dir)
	Elevator.Behaviour = Moving
	Elevator.Dir = dir
}

func onFloorArrival(newFloor int) {
	Elevator.Floor = newFloor
	elevio.SetFloorIndicator(newFloor)

	if scheduler.ShouldStop(Elevator.Floor, Elevator.Dir) {
		elevio.SetMotorDirection(definitions.Stop)
		elevio.SetDoorOpenLamp(true)
		scheduler.ClearOrders(Elevator.Floor, Elevator.Dir)
		doorTimerResetCh <- true
		setAllLights()
		Elevator.Behaviour = DoorOpen
	}
}

func onDoorTimeout() {
	Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(Elevator.Dir)
	if Elevator.Dir == definitions.Stop {
		Elevator.Behaviour = Idle
	} else {
		Elevator.Behaviour = Moving
	}
}

func doorTimer(resetCh chan bool) {
	timer := time.NewTimer(3 * time.Second)
	timer.Stop()
	for {
		select {
		case <-resetCh:
			timer.Reset(3 * time.Second)
		case <-timer.C:
			onDoorTimeout()
		}
	}
}

func setAllLights() {
	for floor := 0; floor < definitions.NumFloors; floor++ {
		for btn := definitions.ButtonType(0); btn < definitions.NumButtons; btn++ {
			elevio.SetButtonLamp(btn, floor, ordermanager.GetOrder(floor, btn))
		}
	}
}
