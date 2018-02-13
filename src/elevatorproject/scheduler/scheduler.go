package scheduler

import (
	"elevatorproject/definitions"
	"elevatorproject/ordermanager"
)

func ShouldStop(floor int, dir definitions.Direction) bool {
	return true
}

func ClearOrders(floor int, dir definitions.Direction) {
	return
}

func ChooseDirection(floor int, dir definitions.Direction) definitions.Direction {
	switch dir {
	case definitions.Up:
		if ordermanager.OrdersAbove(floor) {
			return definitions.Up
		} else if ordermanager.OrdersBelow(floor) {
			return definitions.Down
		} else {
			return definitions.Stop
		}
	case definitions.Down, definitions.Stop:
		if ordermanager.OrdersBelow(floor) {
			return definitions.Down
		} else if ordermanager.OrdersAbove(floor) {
			return definitions.Up
		} else {
			return definitions.Stop
		}
	}
	return definitions.Stop // should never happen
}
