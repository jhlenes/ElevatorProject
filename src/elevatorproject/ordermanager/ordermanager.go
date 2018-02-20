package ordermanager

import (
	def "elevatorproject/definitions"
)

var matrix [def.NumFloors][def.NumButtons]int

func HasOrder(floor int, button def.ButtonType) bool {
	return matrix[floor][button] > 0
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
	matrix[floor][button] = 0
}

func AddOrder(floor int, button def.ButtonType) {
	matrix[floor][button] = 1
}
