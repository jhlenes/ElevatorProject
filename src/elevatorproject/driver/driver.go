package driver

import (
	"log"
	"net"
	"sync"
	"time"
)

type MotorDirection int

const (
	MD_Down MotorDirection = -1
	MD_Stop MotorDirection = 0
	MD_Up   MotorDirection = 1
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown ButtonType = 1
	BT_Cab      ButtonType = 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

const pollRate = 20 * time.Millisecond
const buttonCount = 3

var initialized bool = false
var floorCount int = 4
var mutex sync.Mutex
var conn net.Conn

func Init(addr string, numFloors int) {
	if initialized {
		log.Println("Driver already initialized!")
		return
	}
	floorCount = numFloors
	mutex = sync.Mutex{}
	var err error
	conn, err = net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("ERROR: Could not connect to elevator at: %v\n > %v", addr, err.Error())
	}

	initialized = true
}

func SetMotorDirection(dir MotorDirection) {
	mutex.Lock()
	defer mutex.Unlock()
	conn.Write([]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button ButtonType, floor int, value bool) {
	mutex.Lock()
	defer mutex.Unlock()
	conn.Write([]byte{2, byte(button), byte(floor), toByte(value)})
}

func SetFloorIndicator(floor int) {
	mutex.Lock()
	defer mutex.Unlock()
	conn.Write([]byte{3, byte(floor), 0, 0})
}

func SetDoorOpenLamp(value bool) {
	mutex.Lock()
	defer mutex.Unlock()
	conn.Write([]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	mutex.Lock()
	defer mutex.Unlock()
	conn.Write([]byte{5, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, floorCount)
	for {
		time.Sleep(pollRate)
		for f := 0; f < floorCount; f++ {
			for b := ButtonType(0); b < buttonCount; b++ {
				v := getButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- ButtonEvent{f, ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(pollRate)
		v := getFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(pollRate)
		v := getStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(pollRate)
		v := getObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func getButton(button ButtonType, floor int) bool {
	mutex.Lock()
	defer mutex.Unlock()
	conn.Write([]byte{6, byte(button), byte(floor), 0})
	var buf [4]byte
	conn.Read(buf[:])
	return toBool(buf[1])
}

func getFloor() int {
	mutex.Lock()
	defer mutex.Unlock()
	conn.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	conn.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	} else {
		return -1
	}
}

func getStop() bool {
	mutex.Lock()
	defer mutex.Unlock()
	conn.Write([]byte{8, 0, 0, 0})
	var buf [4]byte
	conn.Read(buf[:])
	return toBool(buf[1])
}

func getObstruction() bool {
	mutex.Lock()
	defer mutex.Unlock()
	conn.Write([]byte{9, 0, 0, 0})
	var buf [4]byte
	conn.Read(buf[:])
	return toBool(buf[1])
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
