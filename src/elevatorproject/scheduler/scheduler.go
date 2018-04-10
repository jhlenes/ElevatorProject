package scheduler

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	om "elevatorproject/ordermanager"
)

func ShouldStop(floor int, dir driver.MotorDirection) bool {
	orders := om.GetOrders(def.LocalId)
	return shouldStop(floor, dir, orders)
}

func ShouldOpenDoor(floor int) bool {
	return om.GetOrders(def.LocalId).HasOrderOnFloor(floor)
}

func ClearOrders(floor int, dir driver.MotorDirection) {
	orders := om.GetOrders(def.LocalId)
	clearOrders(floor, dir, orders)
}

func ChooseDirection(floor int, dir driver.MotorDirection) driver.MotorDirection {
	orders := om.GetOrders(def.LocalId)
	return chooseDirection(floor, dir, orders)
}

func AddOrder(elevator def.Elevator, floor int, button driver.ButtonType) {
	orders := om.GetOrders(def.LocalId)
	if !om.ButtonPressed(floor, button) {
		if button == driver.BT_Cab {
			orders.AddCabOrder(floor, def.LocalId)
		} else {
			cost := timeToIdle(elevator, orders, floor, button)
			orders.AddOrder(floor, button, cost)
		}
	}
}

func AddOrderWithOwner(elevator def.Elevator, floor int, button driver.ButtonType, owner int) {
	AddOrder(elevator, floor, button)
	om.GetOrders(def.LocalId).SetOwner(floor, button, owner)
}

func timeToIdle(elev def.Elevator, orders om.Orders, floor int, button driver.ButtonType) int {
	if elev.Stuck {
		return -1
	}

	// make a copy
	var ordersCopy om.Orders
	orderMatrixValue := *orders.(*om.OrderMatrix) // cast to *OrderMatrix and then get the OrderMatrix
	ordersCopy = &orderMatrixValue

	// add order to local copy of orderMatrix
	ordersCopy.SetStatus(floor, button, om.OS_Existing)
	ordersCopy.SetOwner(floor, button, def.LocalId)

	arrivedAtRequest := false
	duration := 0

	switch elev.Behaviour {
	case def.Idle:
		elev.Dir = chooseDirection(elev.Floor, elev.Dir, ordersCopy)
		if elev.Dir == driver.MD_Stop {
			return duration*10 + def.LocalId
		}
	case def.Moving:
		duration += def.TRAVEL_TIME / 2
		elev.Floor += int(elev.Dir)
	case def.DoorOpen:
		duration += def.DoorTimeout / 2
		elev.Dir = chooseDirection(elev.Floor, elev.Dir, ordersCopy)
	}

	for {
		if shouldStop(elev.Floor, elev.Dir, ordersCopy) {
			clearOrders(elev.Floor, elev.Dir, ordersCopy)
			arrivedAtRequest = !ordersCopy.HasOrder(floor, button)
			if arrivedAtRequest {
				return duration*10 + def.LocalId
			}
			duration += def.DoorTimeout
			elev.Dir = chooseDirection(elev.Floor, elev.Dir, ordersCopy)
		} else if elev.Dir == driver.MD_Stop {
			if ordersCopy.HasOrder(elev.Floor, driver.BT_HallUp) {
				elev.Dir = driver.MD_Up
				clearOrders(elev.Floor, elev.Dir, ordersCopy)
			} else if ordersCopy.HasOrder(elev.Floor, driver.BT_HallDown) {
				elev.Dir = driver.MD_Down
				clearOrders(elev.Floor, elev.Dir, ordersCopy)
			} else if ordersCopy.HasOrder(elev.Floor, driver.BT_Cab) {
				clearOrders(elev.Floor, elev.Dir, ordersCopy)
			}
		}

		elev.Floor += int(elev.Dir)
		duration += def.TRAVEL_TIME
	}
}

func AddCosts(elev def.Elevator) {
	orders := om.GetOrders(def.LocalId)
	for floor := 0; floor < def.FloorCount; floor++ {
		for button := driver.ButtonType(0); button < def.ButtonCount; button++ {
			if om.ButtonPressed(floor, button) {
				orders.SetCost(floor, button, timeToIdle(elev, orders, floor, button))
			}
		}
	}
}

func RemoveCosts() {
	orders := om.GetOrders(def.LocalId)
	for floor := 0; floor < def.FloorCount; floor++ {
		for button := driver.ButtonType(0); button < def.ButtonCount; button++ {
			orders.SetCost(floor, button, -1)
		}
	}
}


func StealOrder() {
	orders := om.GetOrders(def.LocalId)
	for floor := 0; floor < def.FloorCount; floor++ {
		for button := driver.ButtonType(0); button < def.ButtonCount; button++ {
			if orders.GetStatus(floor, button) == om.OS_Existing && button != driver.BT_Cab {
				orders.SetOwner(floor, button, def.LocalId)
				return
			}
		}
	}
}


func shouldStop(floor int, dir driver.MotorDirection, orderMatrix om.Orders) bool {
	switch dir {
	case driver.MD_Down:
		return orderMatrix.HasOrder(floor, driver.BT_Cab) ||
			orderMatrix.HasOrder(floor, driver.BT_HallDown) ||
			!orderMatrix.HasOrderBelow(floor)
	case driver.MD_Up:
		return orderMatrix.HasOrder(floor, driver.BT_Cab) ||
			orderMatrix.HasOrder(floor, driver.BT_HallUp) ||
			!orderMatrix.HasOrderAbove(floor)
	}
	return false
}

func clearOrders(floor int, dir driver.MotorDirection, orders om.Orders) {
	updateOrder(floor, driver.BT_Cab, orders)

	switch dir {
	case driver.MD_Down:
		updateOrder(floor, driver.BT_HallDown, orders)
		if !orders.HasOrderBelow(floor) {
			updateOrder(floor, driver.BT_HallUp, orders)
		}
	case driver.MD_Up:
		updateOrder(floor, driver.BT_HallUp, orders)
		if !orders.HasOrderAbove(floor) {
			updateOrder(floor, driver.BT_HallDown, orders)
		}
	}
}

func updateOrder(floor int, button driver.ButtonType, orders om.Orders) {
	if !orders.HasOrder(floor, button) {
		return
	}
	if button == driver.BT_Cab {
		orders.RemoveOrder(floor, button)
	} else {
		orders.SetStatus(floor, button, om.OS_Completed)
	}
}

func chooseDirection(floor int, dir driver.MotorDirection, orderMatrix om.Orders) driver.MotorDirection {
	switch dir {
	case driver.MD_Up:
		if orderMatrix.HasOrderAbove(floor) {
			return driver.MD_Up
		} else if orderMatrix.HasOrderBelow(floor) {
			return driver.MD_Down
		}
	case driver.MD_Down, driver.MD_Stop:
		if orderMatrix.HasOrderBelow(floor) {
			return driver.MD_Down
		} else if orderMatrix.HasOrderAbove(floor) {
			return driver.MD_Up
		}
	}
	return driver.MD_Stop
}
