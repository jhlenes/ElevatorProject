package fsm

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/ordermanager"
	"elevatorproject/scheduler"
	"time"
)

// Channels used to reset timers
var doorTimerResetCh chan bool = make(chan bool, 10)
var watchdogTimerResetCh chan bool = make(chan bool, 10)

func onDoorTimeout() {
	resetWatchdogTimer()

	Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
	driver.SetDoorOpenLamp(false)
	if Elevator.Dir == driver.MD_Stop {
		Elevator.Behaviour = def.Idle
	} else {
		Elevator.Behaviour = def.Moving
		driver.SetMotorDirection(Elevator.Dir)
	}
}

func resetDoorTimer() {
	resetWatchdogTimer()
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
			onDoorTimeout()
		}
	}
}

func onWatchdogTimeout() {
	resetWatchdogTimer()

	switch Elevator.Behaviour {
	case def.Idle:
		Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
		if Elevator.Dir == driver.MD_Stop {
			completeOrdersOnCurrentFloor()
		} else {
			Elevator.Behaviour = def.Moving
			driver.SetMotorDirection(Elevator.Dir)
		}
	case def.DoorOpen:
		// TODO: Figure out if this can happen and what to do
	case def.Moving: // Elevator is stuck
		// TODO: Figure out what to do here
		// try to restart motor
		//driver.SetMotorDirection(Elevator.Dir)
		// broadcast stuck status? let other elevators take its orders
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
			onWatchdogTimeout()
		}
	}
}

func completeOrdersOnCurrentFloor() {
	orderMatrix := ordermanager.GetMatrix(def.LocalID)
	if orderMatrix.HasOrder(Elevator.Floor, driver.BT_HallUp) {
		Elevator.Dir = driver.MD_Up
		scheduler.ClearOrders(Elevator.Floor, Elevator.Dir)
		driver.SetDoorOpenLamp(true)
		Elevator.Behaviour = def.DoorOpen
		resetDoorTimer()
	} else if orderMatrix.HasOrder(Elevator.Floor, driver.BT_HallDown) {
		Elevator.Dir = driver.MD_Down
		scheduler.ClearOrders(Elevator.Floor, Elevator.Dir)
		driver.SetDoorOpenLamp(true)
		Elevator.Behaviour = def.DoorOpen
		resetDoorTimer()
	} else if orderMatrix.HasOrder(Elevator.Floor, driver.BT_Cab) {
		scheduler.ClearOrders(Elevator.Floor, driver.MD_Stop)
		driver.SetDoorOpenLamp(true)
		Elevator.Behaviour = def.DoorOpen
		resetDoorTimer()
	}
}
