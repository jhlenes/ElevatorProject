package ordermanager

import (
	def "elevatorproject/definitions"
	"math/rand"
	"time"
)

// maybe have a lock?
var matrices [def.NumElevators]def.Matrix

func Init() {
	// Create empty order matrix
	m := def.Matrix{}
	for f := 0; f < def.NumFloors; f++ {
		for b := def.ButtonType(0); b < def.NumButtons; b++ {
			m[f][b] = def.NewOrder()
		}
	}
	matrices[def.LocalID] = m

	rand.Seed(time.Now().UTC().UnixNano())
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
	matrices[def.LocalID][floor][button] = def.NewOrder()
}

func UpdateOrder(floor int, button def.ButtonType) {
	matrices[def.LocalID][floor][button].Status = 2
}

func AddOrder(floor int, button def.ButtonType) {
	matrices[def.LocalID][floor][button].Status = 1
	// TODO: create cost function
	matrices[def.LocalID][floor][button].Cost = rand.Intn(10)*10 + def.LocalID
}

func AddMatrix(id int, newMatrix def.Matrix) {
	matrices[id] = newMatrix
	Synchronise(id)
}

func Synchronise(id int) {
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

			} else if matrices[def.LocalID][f][b].Status == matrices[id][f][b].Status { // equal status

				if matrices[id][f][b].Status == 1 && matrices[def.LocalID][f][b].Owner < 0 { // both status 1 => decide owner
					if matrices[id][f][b].Owner >= 0 { // copy owner
						matrices[def.LocalID][f][b].Owner = matrices[id][f][b].Owner
					} else { // decide owner
						if matrices[def.LocalID][f][b].Cost < matrices[id][f][b].Cost {
							matrices[def.LocalID][f][b].Owner = def.LocalID
						} else {
							matrices[def.LocalID][f][b].Owner = id
						}
					}
					if matrices[def.LocalID][f][b].Owner == def.LocalID {
						// TODO: notify elevator of new orders
						def.Info.Print("We have a new order")
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
		ordersUpdate <- matrices[def.LocalID]
	}
}
