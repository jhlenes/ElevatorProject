package main

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/ordermanager"
	"elevatorproject/scheduler"
	"time"
)

// Channels used to reset timers
var doorTimerResetCh chan bool = make(chan bool)
var watchdogTimerResetCh chan bool = make(chan bool)

func onDoorTimeout() {
	def.Info.Printf("Door timedout.")
	resetWatchdogTimer()

	Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
	driver.SetDoorOpenLamp(false)
	driver.SetMotorDirection(Elevator.Dir)
	if Elevator.Dir == def.Stop {
		Elevator.Behaviour = def.Idle
	} else {
		Elevator.Behaviour = def.Moving
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

func onWatchdogTimeout() {
	resetWatchdogTimer()
	setAllLights() // TODO: should not be here

	switch Elevator.Behaviour {
	case def.Idle:
		Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
		driver.SetMotorDirection(Elevator.Dir)
		if Elevator.Dir == def.Stop {
			moveIfIdleAndHasOrderOnCurrentFloor()
		} else {
			Elevator.Behaviour = def.Moving
		}
	case def.DoorOpen:
		// TODO: Figure out if this can happen and what to do
		onDoorTimeout()
	case def.Moving: // Elevator is stuck
		// TODO: Figure out what to do here
		// try to restart motor
		driver.SetMotorDirection(Elevator.Dir)
		// broadcast stuck status? let other elevators take its orders
	}
}

func resetWatchdogTimer() {
	// TODO: sync button lights here?
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

func moveIfIdleAndHasOrderOnCurrentFloor() {
	if ordermanager.HasOrder(Elevator.Floor, def.BT_HallUp) {
		Elevator.Dir = def.Up
		scheduler.ClearOrders(Elevator.Floor, Elevator.Dir)
		driver.SetDoorOpenLamp(true)
		Elevator.Behaviour = def.DoorOpen
		resetDoorTimer()
	} else if ordermanager.HasOrder(Elevator.Floor, def.BT_HallDown) {
		Elevator.Dir = def.Down
		scheduler.ClearOrders(Elevator.Floor, Elevator.Dir)
		driver.SetDoorOpenLamp(true)
		Elevator.Behaviour = def.DoorOpen
		resetDoorTimer()
	} else if ordermanager.HasOrder(Elevator.Floor, def.BT_Cab) {
		scheduler.ClearOrders(Elevator.Floor, def.Stop)
		driver.SetDoorOpenLamp(true)
		Elevator.Behaviour = def.DoorOpen
		resetDoorTimer()
	}
}
