package scheduler

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/ordermanager"
	"fmt"
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
	cost := CalculateCost(elevator, button, floor)
	orderMatrix.AddOrder(floor, button, cost)
	fmt.Printf("Cost: %v\n", cost)
}

func CalculateCost(eOld def.Elevator, button driver.ButtonType, floor int) int {
	e := eOld
	ordersCopy := *ordermanager.GetLocalOrderMatrix()
	ordersCopy[floor][button].Status = 1
	ordersCopy[floor][button].Owner = def.LocalID

	arrivedAtRequest := false
	duration := 0

	switch e.Behaviour {
	case def.Idle:
		e.Dir = chooseDirection(e.Floor, e.Dir, &ordersCopy)
		if e.Dir == driver.MD_Stop {
			return duration*10 + def.LocalID
		}
	case def.Moving:
		duration += def.TRAVEL_TIME / 2
		e.Floor += int(e.Dir)
	case def.DoorOpen:
		duration -= def.DoorTimeout / 2
	}

	for {
		if shouldStop(e.Floor, e.Dir, &ordersCopy) {
			clearOrders(e.Floor, e.Dir, &ordersCopy)
			arrivedAtRequest = !ordersCopy.HasOrder(floor, button)
			if arrivedAtRequest {
				return duration*10 + def.LocalID
			}
			duration += def.DoorTimeout
			e.Dir = chooseDirection(e.Floor, e.Dir, &ordersCopy)
		}
		e.Floor += int(e.Dir)
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
