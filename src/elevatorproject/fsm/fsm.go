package fsm

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/ordermanager"
	"elevatorproject/scheduler"
	"fmt"
)

var Elevator def.Elevator

func Init() {
	// Initialize driver
	elevatorAddr := fmt.Sprintf("%s:%d", def.Addr, def.Port)
	driver.Init(elevatorAddr, def.FloorCount)

	Elevator.Floor = -1
	Elevator.Dir = driver.MD_Up
	driver.SetMotorDirection(driver.MD_Up)
	Elevator.Behaviour = def.Initializing
	Elevator.ID = def.LocalID

	go doorTimer(doorTimerResetCh)
	go watchdogTimer(watchdogTimerResetCh)
	go listenForDriverEvents()
}

func listenForDriverEvents() {

	// Create channels and start polling events
	drvButtons := make(chan driver.ButtonEvent, 10)
	drvFloors := make(chan int)
	drvObstr := make(chan bool)
	drvStop := make(chan bool)
	go driver.PollButtons(drvButtons)
	go driver.PollFloorSensor(drvFloors)
	go driver.PollObstructionSwitch(drvObstr)
	go driver.PollStopButton(drvStop)

	// Listen for events at the channels
	for {
		select {
		case button := <-drvButtons:
			def.Info.Println(button)
			go onButtonPress(button)

		case floor := <-drvFloors:
			onFloorArrival(floor)

		case a := <-drvObstr:
			def.Info.Printf("Obstruction: %+v\n", a)

		case a := <-drvStop:
			def.Info.Printf("Stop: %+v\n", a)
		}
	}
}

func OnNewOrder(floor int, button driver.ButtonType) {
	resetWatchdogTimer()

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
			driver.SetMotorDirection(Elevator.Dir)
			Elevator.Behaviour = def.Moving
		}
	}

}

func onButtonPress(buttonEvent driver.ButtonEvent) {
	resetWatchdogTimer()

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
		}
	}

	if !orderCompleted {
		scheduler.AddOrder(Elevator, buttonEvent.Floor, buttonEvent.Button)
	}
}

func onFloorArrival(newFloor int) {
	resetWatchdogTimer()

	if Elevator.Behaviour == def.Initializing {
		Elevator.Behaviour = def.Idle
		Elevator.Dir = driver.MD_Stop
		def.Info.Printf("Done initializing")
	}

	driver.SetMotorDirection(Elevator.Dir) // make sure elevator is going the way is says it is

	Elevator.Floor = newFloor
	driver.SetFloorIndicator(newFloor)
	if scheduler.ShouldStop(Elevator.Floor, Elevator.Dir) {
		driver.SetMotorDirection(driver.MD_Stop)
		driver.SetDoorOpenLamp(true)
		scheduler.ClearOrders(Elevator.Floor, Elevator.Dir)
		resetDoorTimer()
		Elevator.Behaviour = def.DoorOpen
	}
}

func SetAllLights(elevatorIds []int) {
	for floor := 0; floor < def.FloorCount; floor++ {
		for btn := driver.ButtonType(0); btn < def.ButtonCount; btn++ {
			driver.SetButtonLamp(btn, floor, ordermanager.HasSystemOrder(floor, btn, elevatorIds))
		}
	}
}
