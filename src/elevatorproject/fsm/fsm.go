package fsm

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	om "elevatorproject/ordermanager"
	"elevatorproject/scheduler"
	"fmt"
)

// We can ignore button presses if there are not enough active elevators
var NumActiveElevators = 1

var Elevator def.Elevator
var buttonStatus [def.FloorCount][def.ButtonCount]bool

func Init() {

	// Initialize dependencies
	elevatorAddr := fmt.Sprintf("%s:%d", def.Addr, def.Port)
	driver.Init(elevatorAddr, def.FloorCount)
	om.ReadBackup()

	// Initialize elevator states
	Elevator.Floor = -1
	Elevator.Dir = driver.MD_Up
	driver.SetMotorDirection(driver.MD_Up)
	Elevator.Behaviour = def.Initializing
	Elevator.ID = def.LocalId

	// Set all lights to their correct state
	for floor := 0; floor < def.FloorCount; floor++ {
		for button := driver.ButtonType(0); button < def.ButtonCount; button++ {
			driver.SetButtonLamp(button, floor, false)
		}
	}
	driver.SetDoorOpenLamp(false)
	SetAllLights()

	go doorTimer(doorTimerResetCh)
	go watchdogTimer(watchdogTimerResetCh)
	go listenForDriverEvents()
}

func listenForDriverEvents() {

	// Create channels and start polling events
	drvButtons := make(chan driver.ButtonEvent, 10)
	drvFloors := make(chan int, 10)
	go driver.PollButtons(drvButtons)
	go driver.PollFloorSensor(drvFloors)

	// Listen for events at the channels
	for {
		select {
		case button := <-drvButtons:
			if Elevator.Behaviour != def.Initializing {
				go onButtonPress(button)
			}

		case floor := <-drvFloors:
			go onFloorArrival(floor)
		}
	}
}

// OnNewOrder should be called when a new order has been assigned to this elevator.
// If idle or door open, this function starts the motor in the direction of the new order or opens the door if the order is on the same floor.
func OnNewOrder(floor int, button driver.ButtonType) {
	switch Elevator.Behaviour {
	case def.DoorOpen:
		if Elevator.Floor == floor {

			// if door is open, at correct floor and going in button direction => clear order
			if button == driver.BT_HallUp && Elevator.Dir != driver.MD_Down {
				resetDoorTimer()
				scheduler.ClearOrders(floor, driver.MD_Up)
			} else if button == driver.BT_HallDown && Elevator.Dir != driver.MD_Up {
				resetDoorTimer()
				scheduler.ClearOrders(floor, driver.MD_Down)
			}
		}
	case def.Idle:
		if Elevator.Floor == floor {
			driver.SetDoorOpenLamp(true)
			Elevator.Behaviour = def.DoorOpen
			resetDoorTimer()
			scheduler.ClearOrders(Elevator.Floor, driver.MD_Up)
			scheduler.ClearOrders(Elevator.Floor, driver.MD_Down)
		} else {
			Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
			if Elevator.Dir != driver.MD_Stop {
				driver.SetMotorDirection(Elevator.Dir)
				Elevator.Behaviour = def.Moving
				resetWatchdogTimer()
			}
		}
	}

}

// onButtonPress adds the order to the ordermanager if enough active elevators, or completes the order right away if it is on the same floor
func onButtonPress(buttonEvent driver.ButtonEvent) {
	orderCompleted := false
	switch Elevator.Behaviour {
	case def.DoorOpen:
		if Elevator.Floor == buttonEvent.Floor {
			resetDoorTimer()
			orderCompleted = true
		}
	case def.Idle:
		if Elevator.Floor == buttonEvent.Floor {
			driver.SetDoorOpenLamp(true)
			Elevator.Behaviour = def.DoorOpen
			resetDoorTimer()
			orderCompleted = true
		} else if buttonEvent.Button == driver.BT_Cab {

			// We can start a cab order without confirmation from other elevators
			scheduler.AddOrder(Elevator, buttonEvent.Floor, buttonEvent.Button)
			Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
			driver.SetMotorDirection(Elevator.Dir)
			Elevator.Behaviour = def.Moving
			orderCompleted = true
			resetWatchdogTimer()
		}
	}

	if NumActiveElevators < 2 && buttonEvent.Button != driver.BT_Cab { // Ignore some button presses when only 1 elevator
		return
	}

	if !orderCompleted {
		scheduler.AddOrder(Elevator, buttonEvent.Floor, buttonEvent.Button)
	}

	// cab orders are accepted immedeately, so we need to set the lights accordingly
	if buttonEvent.Button == driver.BT_Cab {
		SetAllLights()
	}
}

func onFloorArrival(newFloor int) {
	resetWatchdogTimer()

	if Elevator.Behaviour == def.Initializing {
		Elevator.Behaviour = def.Idle
		Elevator.Dir = driver.MD_Stop
		driver.SetMotorDirection(Elevator.Dir)
	} else if Elevator.Stuck {
		Elevator.Stuck = false
		scheduler.AddCosts(Elevator)
	}

	Elevator.Floor = newFloor
	driver.SetFloorIndicator(newFloor)
	if scheduler.ShouldStop(Elevator.Floor, Elevator.Dir) {

		if scheduler.ShouldOpenDoor(newFloor) {
			driver.SetMotorDirection(driver.MD_Stop)
			driver.SetDoorOpenLamp(true)
			scheduler.ClearOrders(Elevator.Floor, Elevator.Dir)
			Elevator.Behaviour = def.DoorOpen
			resetDoorTimer()
			SetAllLights()
		} else {
			Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
			driver.SetMotorDirection(Elevator.Dir)
			if Elevator.Dir == driver.MD_Stop {
				Elevator.Behaviour = def.Idle
			}
		}
	}
}

func SetAllLights() {
	for floor := 0; floor < def.FloorCount; floor++ {
		for button := driver.ButtonType(0); button < def.ButtonCount; button++ {
			SetLight(floor, button)
		}
	}
}

// SetLight tells the driver to change the button lamp if necessary
func SetLight(floor int, button driver.ButtonType) {
	if button == driver.BT_Cab {
		if bStatus := om.GetOrders(def.LocalId).HasOrder(floor, button); bStatus != buttonStatus[floor][button] {
			driver.SetButtonLamp(button, floor, bStatus)
			buttonStatus[floor][button] = bStatus
		}
	} else {
		if bStatus := om.GetOrders(def.LocalId).HasSystemOrder(floor, button); bStatus != buttonStatus[floor][button] {
			driver.SetButtonLamp(button, floor, bStatus)
			buttonStatus[floor][button] = bStatus
		}
	}
}
