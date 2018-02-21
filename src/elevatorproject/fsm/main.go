package main

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/network"
	"elevatorproject/ordermanager"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
)

func main() {
	initFsm()
	driver.Init(fmt.Sprintf("%s%d", def.Addr, def.Port), def.NumFloors)
	ordermanager.Init()
	network.Init()

	// Create channels and start polling events
	drvButtons := make(chan def.ButtonEvent)
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
			onButtonPress(button)

		case floor := <-drvFloors:
			onFloorArrival(floor)

		case a := <-drvObstr:
			def.Info.Printf("Obstruction: %+v\n", a)
			if a {
				driver.SetMotorDirection(def.Stop)
			} else {
				driver.SetMotorDirection(Elevator.Dir)
			}

		case a := <-drvStop:
			def.Info.Printf("Stop: %+v\n", a)
			/*for f := 0; f < def.NumFloors; f++ {
				for b := def.ButtonType(0); b < 3; b++ {
					driver.SetButtonLamp(b, f, false)
				}
			}*/
		}
	}
}

func initFsm() {

	Elevator.Floor = -1
	Elevator.Dir = def.Stop
	Elevator.Behaviour = def.Idle

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	i, err := strconv.Atoi(id)
	if err != nil {
		def.Error.Fatalf("ID is not an integer: %v", id)
	} else {
		def.Port += i
		def.LocalID = i
	}

	Elevator.ID = def.LocalID

	go safeShutdown()

	// start timers
	go doorTimer(doorTimerResetCh)
	go watchdogTimer(watchdogTimerResetCh)
}

// safeShutdown shutdowns the program in a safe way when terminataed by user (ctrl+c)
func safeShutdown() {
	var c = make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
	driver.SetMotorDirection(def.Stop)
	def.Info.Println("User terminated the program.")
	os.Exit(1)
}
