package ordermanager

import (
	def "elevatorproject/definitions"
	"time"
)

// maybe have a lock
var matrices map[string]*def.Matrix = make(map[string]*def.Matrix)

func Init() {
	matrices[def.LocalID] = &def.Matrix{}
}

func HasOrder(floor int, button def.ButtonType) bool {
	return matrices[def.LocalID][floor][button].Owner == def.LocalID
}

func HasOrderAbove(floor int) bool {
	for f := floor + 1; f < def.NumFloors; f++ {
		for b := def.ButtonType(0); b < def.NumButtons; b++ {
			if HasOrder(f, b) {
				return true
			}
		}
	}
	return false
}

func HasOrderBelow(floor int) bool {
	for f := 0; f < floor; f++ {
		for b := def.ButtonType(0); b < def.NumButtons; b++ {
			if HasOrder(f, b) {
				return true
			}
		}
	}
	return false
}

func RemoveOrder(floor int, button def.ButtonType) {
	matrices[def.LocalID][floor][button].Status = 0
	matrices[def.LocalID][floor][button].Owner = ""
	matrices[def.LocalID][floor][button].Cost = -1
}

func AddOrder(floor int, button def.ButtonType) {
	matrices[def.LocalID][floor][button].Status = 1
	// add cost
}

func AddMatrix(id string, newMatrix def.Matrix) {
	*matrices[id] = newMatrix
	Synchronise(id)
}

func Synchronise(id string) {
	for f := 0; f < def.NumFloors; f++ {
		for b := def.ButtonType(0); b < def.NumButtons; b++ {

			// assuming 2 elevators
			if matrices[def.LocalID][f][b].Status < matrices[id][f][b].Status { // update status to higher number

				if matrices[id][f][b].Status == 1 { // new order
					AddOrder(f, b)
				} else if matrices[id][f][b].Status == 2 { // finished order
					matrices[def.LocalID][f][b].Status = 2
				} else if matrices[def.LocalID][f][b].Status != 0 && matrices[id][f][b].Status == 3 { // deleting order
					matrices[def.LocalID][f][b].Status = 3
				}

			} else if matrices[id][f][b].Status == matrices[id][f][b].Status { // equal status

				if matrices[id][f][b].Status == 1 && matrices[def.LocalID][f][b].Owner == "" { // both status 1 => decide owner
					if matrices[id][f][b].Owner != "" { // copy owner
						matrices[def.LocalID][f][b].Owner = matrices[id][f][b].Owner
					} else { // decide owner
						if matrices[def.LocalID][f][b].Cost < matrices[id][f][b].Cost {
							matrices[def.LocalID][f][b].Owner = def.LocalID
						} else {
							matrices[def.LocalID][f][b].Owner = id
						}
					}
				} else if matrices[id][f][b].Status == 2 { //  TODO: check that all have finished order
					matrices[def.LocalID][f][b].Status = 3
				} else if matrices[def.LocalID][f][b].Status == 3 {
					RemoveOrder(f, b)
				}
			} else if matrices[def.LocalID][f][b].Status == 3 && matrices[id][f][b].Status == 0 { //
				RemoveOrder(f, b)
			}
		}
	}
}

func PollOrders(ordersUpdate chan def.Matrix) {
	for {
		time.Sleep(def.SendTime * time.Millisecond)
		ordersUpdate <- *matrices[def.LocalID]
	}
}
