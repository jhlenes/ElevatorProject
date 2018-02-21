package main

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/ordermanager"
	"elevatorproject/scheduler"
)

var Elevator def.Elevator

func onButtonPress(button def.ButtonEvent) {
	def.Info.Printf("New request: %+v", button)
	resetWatchdogTimer()

	ordermanager.AddOrder(button.Floor, button.Button)
	switch Elevator.Behaviour {
	case def.DoorOpen:
		if Elevator.Floor == button.Floor {
			resetDoorTimer()
			ordermanager.RemoveOrder(button.Floor, button.Button)
		}
	case def.Idle:
		if Elevator.Floor == button.Floor {
			driver.SetDoorOpenLamp(true)
			Elevator.Behaviour = def.DoorOpen
			resetDoorTimer()
			ordermanager.RemoveOrder(button.Floor, button.Button)
		}
	}
}

func onFloorArrival(newFloor int) {
	def.Info.Printf("Arrived at floor %v.", newFloor)
	resetWatchdogTimer()

	if Elevator.Floor == -1 { // elevator was initialized between floors
		Elevator.Dir = def.Stop
	}

	driver.SetMotorDirection(Elevator.Dir) // make sure elevator is going the way is says it is

	Elevator.Floor = newFloor
	driver.SetFloorIndicator(newFloor)
	if scheduler.ShouldStop(Elevator.Floor, Elevator.Dir) {
		driver.SetMotorDirection(def.Stop)
		driver.SetDoorOpenLamp(true)
		scheduler.ClearOrders(Elevator.Floor, Elevator.Dir)
		resetDoorTimer()
		Elevator.Behaviour = def.DoorOpen
		setAllLights()
	}
}

func setAllLights() {
	for floor := 0; floor < def.NumFloors; floor++ {
		for btn := def.ButtonType(0); btn < def.NumButtons; btn++ {
			driver.SetButtonLamp(btn, floor, ordermanager.HasOrder(floor, btn))
		}
	}
}
