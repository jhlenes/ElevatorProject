package definitions

import (
	"log"
	"os"
)

var LocalID int = 0

const NumFloors = 4
const NumButtons = 3
const NumElevators = 3

const Addr = "localhost:"

var Port = 15657

const DoorTimeout = 3000     // ms
const WatchdogTimeout = 5000 // ms
const SendTime = 2000        // ms

// Loggers
var Info = log.New(os.Stdout,
	"INFO: ",
	log.Ltime)
var Error = log.New(os.Stderr,
	"ERROR: ",
	log.Ltime|log.Lshortfile)

type Direction int

const (
	Up   Direction = 1
	Down           = -1
	Stop           = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type order struct {
	Status int
	Cost   int
	Owner  int
}

func NewOrder() order {
	return order{0, 9999, -1}
}

type Matrix [NumFloors][NumButtons]order

type ElevatorBehaviour int

// Elevator behaviours
const (
	Idle ElevatorBehaviour = iota
	DoorOpen
	Moving
)

type Elevator struct {
	Floor     int
	Dir       Direction
	Behaviour ElevatorBehaviour
	ID        int
}
