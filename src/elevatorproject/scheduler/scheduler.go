package scheduler

import (
	def "elevatorproject/definitions"
	"elevatorproject/ordermanager"
)

func ShouldStop(floor int, dir def.Direction) bool {
	switch dir {
	case def.Down:
		return ordermanager.HasOrder(floor, def.BT_Cab) ||
			ordermanager.HasOrder(floor, def.BT_HallDown) ||
			!ordermanager.HasOrderBelow(floor)
	case def.Up:
		return ordermanager.HasOrder(floor, def.BT_Cab) ||
			ordermanager.HasOrder(floor, def.BT_HallUp) ||
			!ordermanager.HasOrderAbove(floor)
	}
	return false
}

func ClearOrders(floor int, dir def.Direction) {
	ordermanager.UpdateOrder(floor, def.BT_Cab)
	switch dir {
	case def.Down:
		ordermanager.UpdateOrder(floor, def.BT_HallDown)
		if !ordermanager.HasOrderBelow(floor) {
			ordermanager.UpdateOrder(floor, def.BT_HallUp)
		}
	case def.Up:
		ordermanager.UpdateOrder(floor, def.BT_HallUp)
		if !ordermanager.HasOrderAbove(floor) {
			ordermanager.UpdateOrder(floor, def.BT_HallDown)
		}
	}
}

func ChooseDirection(floor int, dir def.Direction) def.Direction {
	switch dir {
	case def.Up:
		if ordermanager.HasOrderAbove(floor) {
			return def.Up
		} else if ordermanager.HasOrderBelow(floor) {
			return def.Down
		}
	case def.Down, def.Stop:
		if ordermanager.HasOrderBelow(floor) {
			return def.Down
		} else if ordermanager.HasOrderAbove(floor) {
			return def.Up
		}
	}
	return def.Stop
}
