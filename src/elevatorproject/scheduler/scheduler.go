package scheduler

import (
	def "elevatorproject/definitions"
	"elevatorproject/ordermanager"
)

func ShouldStop(floor int, dir def.Direction) bool {
	return true
}

func ClearOrders(floor int, dir def.Direction) {
	return
}

func ChooseDirection(floor int, dir def.Direction) def.Direction {
	switch dir {
	case def.Up:
		if ordermanager.OrdersAbove(floor) {
			return def.Up
		} else if ordermanager.OrdersBelow(floor) {
			return def.Down
		}
	case def.Down, def.Stop:
		if ordermanager.OrdersBelow(floor) {
			return def.Down
		} else if ordermanager.OrdersAbove(floor) {
			return def.Up
		}
	}
	return def.Stop
}
