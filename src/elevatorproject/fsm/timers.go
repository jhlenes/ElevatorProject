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
	if (Elevator.Stuck) {
		Elevator.Stuck = false
		scheduler.AddCosts(Elevator)
	}

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
	def.Info.Println("watchdog")
	resetWatchdogTimer()

	switch Elevator.Behaviour {
	case def.Idle:
		if ordermanager.GetOrders(def.LocalId).HasOrderOnFloor(Elevator.Floor) {
			driver.SetDoorOpenLamp(true)
			Elevator.Behaviour = def.DoorOpen
			resetDoorTimer()
			scheduler.ClearOrders(Elevator.Floor, driver.MD_Up)
			scheduler.ClearOrders(Elevator.Floor, driver.MD_Down)
			SetAllLights()
		} else {
			Elevator.Dir = scheduler.ChooseDirection(Elevator.Floor, Elevator.Dir)
			if Elevator.Dir != driver.MD_Stop {
				Elevator.Behaviour = def.Moving
				driver.SetMotorDirection(Elevator.Dir)
			} else {
				if Elevator.Stuck {
					Elevator.Behaviour = def.Moving
					if Elevator.Floor == 0 {
						Elevator.Dir = driver.MD_Up
					} else {
						Elevator.Dir = driver.MD_Down
					}
					driver.SetMotorDirection(Elevator.Dir)
				}
			}
		}
	case def.DoorOpen:
		// TODO: Figure out if this can happen and what to do
		Elevator.Stuck = true
		scheduler.RemoveCosts()
	case def.Moving: // Elevator is stuck
		// TODO: Figure out what to do here
		Elevator.Stuck = true
		scheduler.RemoveCosts()
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
