package scheduler

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/ordermanager"
)

func ShouldStop(floor int, dir driver.MotorDirection) bool {
	orderMatrix := ordermanager.GetOrders(def.LocalID)
	return shouldStop(floor, dir, orderMatrix)
}

func ClearOrders(floor int, dir driver.MotorDirection) {
	orderMatrix := ordermanager.GetOrders(def.LocalID)
	clearOrders(floor, dir, orderMatrix)
}

func ChooseDirection(floor int, dir driver.MotorDirection) driver.MotorDirection {
	orderMatrix := ordermanager.GetOrders(def.LocalID)
	return chooseDirection(floor, dir, orderMatrix)
}

func AddOrder(elevator def.Elevator, floor int, button driver.ButtonType) {
	orderMatrix := ordermanager.GetOrders(def.LocalID)
	if !ordermanager.ButtonPressed(floor, button) {
		if button != driver.BT_Cab {
			cost := timeToIdle(elevator, orderMatrix, floor, button)
			orderMatrix.AddOrder(floor, button, cost)
		} else {
			orderMatrix.AddCabOrder(floor, def.LocalID)
		}
	}
}

func AddOrderWithOwner(elevator def.Elevator, floor int, button driver.ButtonType, owner int) {
	AddOrder(elevator, floor, button)
	ordermanager.GetOrders(def.LocalID).SetOwner(floor, button, owner)
}

func timeToIdle(elev def.Elevator, orders ordermanager.Orders, floor int, button driver.ButtonType) int {
	// make a copy
	var ordersCopy ordermanager.Orders
	orderMatrixValue := *orders.(*ordermanager.OrderMatrix) // cast to *OrderMatrix and then get the OrderMatrix
	ordersCopy = &orderMatrixValue

	// add order to local copy of orderMatrix
	ordersCopy.SetStatus(floor, button, ordermanager.OS_Existing)
	ordersCopy.SetOwner(floor, button, def.LocalID)

	arrivedAtRequest := false
	duration := 0

	switch elev.Behaviour {
	case def.Idle:
		elev.Dir = chooseDirection(elev.Floor, elev.Dir, ordersCopy)
		if elev.Dir == driver.MD_Stop {
			return duration*10 + def.LocalID
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
				return duration*10 + def.LocalID
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

func shouldStop(floor int, dir driver.MotorDirection, orderMatrix ordermanager.Orders) bool {
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

func clearOrders(floor int, dir driver.MotorDirection, orderMatrix ordermanager.Orders) {
	orderMatrix.UpdateOrder(floor, driver.BT_Cab)
	switch dir {
	case driver.MD_Down:
		orderMatrix.UpdateOrder(floor, driver.BT_HallDown)
		if !orderMatrix.HasOrderBelow(floor) {
			orderMatrix.UpdateOrder(floor, driver.BT_HallUp)
		}
	case driver.MD_Up:
		orderMatrix.UpdateOrder(floor, driver.BT_HallUp)
		if !orderMatrix.HasOrderAbove(floor) {
			orderMatrix.UpdateOrder(floor, driver.BT_HallDown)
		}
	}
}

func chooseDirection(floor int, dir driver.MotorDirection, orderMatrix ordermanager.Orders) driver.MotorDirection {
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
