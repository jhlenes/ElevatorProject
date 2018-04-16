package definitions

import (
	"elevatorproject/driver"
	"log"
	"os"
)

var LocalId = 0

// Connection to elevator, default values
var Addr = "localhost"
var Port = 15657

const FloorCount = 4
const ButtonCount = 3
const ElevatorCount = 3

const TravelTime = 2000 // ms
const SendTime = 200    // ms

const DoorTimeout = 3000     // ms
const WatchdogTimeout = 5000 // ms
const NetworkTimeout = 2     // s

// Setup and format logger messages
var Info = log.New(os.Stdout, "INFO: ", log.Ltime)
var Error = log.New(os.Stderr, "ERROR: ", log.Ltime|log.Lshortfile)

type ElevatorBehaviour int

const (
	Idle ElevatorBehaviour = iota
	DoorOpen
	Moving
	Initializing
)

type Elevator struct {
	Floor     int
	Dir       driver.MotorDirection
	Behaviour ElevatorBehaviour
	ID        int
	Stuck     bool
}
