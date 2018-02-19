package definitions

const NumFloors = 4
const NumButtons = 3
const NumElevators = 3

const Addr = "localhost:15657"

const DoorTimeout = 3000     // ms
const WatchdogTimeout = 5000 // ms

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
