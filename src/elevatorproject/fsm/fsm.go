package fsm

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	om "elevatorproject/ordermanager"
	"elevatorproject/scheduler"
	"fmt"
)

var NumOnlineElevators = 0
var Elevator def.Elevator
var buttonStatus [def.FloorCount][def.ButtonCount]bool

func Init() {
	// Initialize driver
	elevatorAddr := fmt.Sprintf("%s:%d", def.Addr, def.Port)
	driver.Init(elevatorAddr, def.FloorCount)
	om.ReadBackup()

	Elevator.Floor = -1
	Elevator.Dir = driver.MD_Up
	driver.SetMotorDirection(driver.MD_Up)
	Elevator.Behaviour = def.Initializing
	Elevator.ID = def.LocalID

	// Reset all lights
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
				def.Info.Printf("%+v\n", button)
				go onButtonPress(button)
			}

		case floor := <-drvFloors:
			go onFloorArrival(floor)
		}
	}
}

func OnNewOrder(floor int, button driver.ButtonType) {
	switch Elevator.Behaviour {
	case def.DoorOpen: // if door is open, at correct floor and going in button direction => clear order
		if Elevator.Floor == floor {
			if button == driver.BT_HallUp && Elevator.Dir == driver.MD_Up {
				resetDoorTimer()
				scheduler.ClearOrders(floor, driver.MD_Up)
			} else if button == driver.BT_HallDown && Elevator.Dir == driver.MD_Down {
				resetDoorTimer()
				scheduler.ClearOrders(floor, driver.MD_Down)
			} else if button == driver.BT_Cab {
				resetDoorTimer()
				scheduler.ClearOrders(floor, driver.MD_Stop)
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
		} else if buttonEvent.Button == driver.BT_Cab && Elevator.Behaviour != def.Stuck {

			// We can start a cab order without confirmation from other elevators
			scheduler.AddOrder(Elevator, buttonEvent.Floor, buttonEvent.Button)
			Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
			driver.SetMotorDirection(Elevator.Dir)
			Elevator.Behaviour = def.Moving
			orderCompleted = true
			resetWatchdogTimer()
		}
	}

	if NumOnlineElevators < 2 && buttonEvent.Button != driver.BT_Cab { // Ignore some button presses when only 1 elevator
		return
	}

	if !orderCompleted {
		scheduler.AddOrder(Elevator, buttonEvent.Floor, buttonEvent.Button)
	}

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
	} else if Elevator.Behaviour == def.Stuck {
		Elevator.Behaviour = def.Moving
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

func SetLight(floor int, button driver.ButtonType) {
	if button == driver.BT_Cab {
		if bStatus := om.GetOrders(def.LocalID).HasOrder(floor, button); bStatus != buttonStatus[floor][button] {
			driver.SetButtonLamp(button, floor, bStatus)
			buttonStatus[floor][button] = bStatus
		}
	} else {
		if bStatus := om.GetOrders(def.LocalID).HasSystemOrder(floor, button); bStatus != buttonStatus[floor][button] {
			driver.SetButtonLamp(button, floor, bStatus)
			buttonStatus[floor][button] = bStatus
		}
	}
}
