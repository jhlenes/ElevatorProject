package scheduler

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/ordermanager"
)

func ShouldStop(floor int, dir driver.MotorDirection) bool {
	orderMatrix := ordermanager.GetLocalOrderMatrix()
	return shouldStop(floor, dir, orderMatrix)
}

func ClearOrders(floor int, dir driver.MotorDirection) {
	orderMatrix := ordermanager.GetLocalOrderMatrix()
	clearOrders(floor, dir, orderMatrix)
}

func ChooseDirection(floor int, dir driver.MotorDirection) driver.MotorDirection {
	orderMatrix := ordermanager.GetLocalOrderMatrix()
	return chooseDirection(floor, dir, orderMatrix)
}

func AddOrder(elevator def.Elevator, floor int, button driver.ButtonType) {
	orderMatrix := ordermanager.GetLocalOrderMatrix()
	if !ordermanager.ButtonPressed(floor, button) {
		if button != driver.BT_Cab {
			cost := timeToIdle(elevator, *orderMatrix, floor, button)
			orderMatrix.AddOrder(floor, button, cost)
		} else {
			orderMatrix.AddCabOrder(floor, def.LocalID)
		}
	}
}

func timeToIdle(elev def.Elevator, orderMatrix ordermanager.OrderMatrix, floor int, button driver.ButtonType) int {
	// add order to local copy of orderMatrix
	orderMatrix[floor][button].Status = 1
	orderMatrix[floor][button].Owner = def.LocalID

	arrivedAtRequest := false
	duration := 0

	switch elev.Behaviour {
	case def.Idle:
		elev.Dir = chooseDirection(elev.Floor, elev.Dir, &orderMatrix)
		if elev.Dir == driver.MD_Stop {
			return duration*10 + def.LocalID
		}
	case def.Moving:
		duration += def.TRAVEL_TIME / 2
		elev.Floor += int(elev.Dir)
	case def.DoorOpen:
		duration -= def.DoorTimeout / 2
		elev.Dir = chooseDirection(elev.Floor, elev.Dir, &orderMatrix)
	}

	for {
		if shouldStop(elev.Floor, elev.Dir, &orderMatrix) {
			clearOrders(elev.Floor, elev.Dir, &orderMatrix)
			arrivedAtRequest = !orderMatrix.HasOrder(floor, button)
			if arrivedAtRequest {
				return duration*10 + def.LocalID
			}
			duration += def.DoorTimeout
			elev.Dir = chooseDirection(elev.Floor, elev.Dir, &orderMatrix)
		} else if elev.Dir == driver.MD_Stop {
			if orderMatrix.HasOrder(elev.Floor, driver.BT_HallUp) {
				elev.Dir = driver.MD_Up
				clearOrders(elev.Floor, elev.Dir, &orderMatrix)
			} else if orderMatrix.HasOrder(elev.Floor, driver.BT_HallDown) {
				elev.Dir = driver.MD_Down
				clearOrders(elev.Floor, elev.Dir, &orderMatrix)
			} else if orderMatrix.HasOrder(elev.Floor, driver.BT_Cab) {
				clearOrders(elev.Floor, elev.Dir, &orderMatrix)
			}
		}

		elev.Floor += int(elev.Dir)
		duration += def.TRAVEL_TIME
	}
}

func shouldStop(floor int, dir driver.MotorDirection, orderMatrix *ordermanager.OrderMatrix) bool {
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

func clearOrders(floor int, dir driver.MotorDirection, orderMatrix *ordermanager.OrderMatrix) {
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

func chooseDirection(floor int, dir driver.MotorDirection, orderMatrix *ordermanager.OrderMatrix) driver.MotorDirection {
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
