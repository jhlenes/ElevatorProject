package synchronizer

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/fsm"
	om "elevatorproject/ordermanager"
	"elevatorproject/scheduler"
)

func ReassignOrders(ids []int, id int) {
	def.Info.Printf("Reassigning orders of %v to elevators: %v\n", id, ids)

	for f := 0; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			if b == driver.BT_Cab {
				continue
			}

			// Reassign orders owned by <id> to the elevator with the lowest cost
			if om.GetOrders(def.LocalID).GetOwner(f, b) == id {
				if cost, bestId := getLowestCost(ids, f, b); cost >= 0 {
					addOrderWithOwner(f, b, bestId)
					if bestId == def.LocalID {
						fsm.OnNewOrder(f, b)
					}
				} else { // couldn't find a new owner
					addOrder(f, b) // TODO: correct?
				}
			}

			// TODO: Should also take order in some cases if communication was lost during confirmation

			om.GetOrders(id).SetOrder(f, b, om.CreateEmptyOrder())
		}
	}
}

func Synchronize(ids []int) {
	for f := 0; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			if b == driver.BT_Cab {
				continue
			}

			// When we are stuck, just copy whatever we receive
			if fsm.Elevator.Behaviour == def.Stuck {
				om.GetOrders(def.LocalID).SetOrder(f, b, om.GetOrders(ids[0]).GetOrder(f, b))
				fsm.SetLight(f, b)
				continue
			}

			switch om.GetOrders(def.LocalID).GetStatus(f, b) {
			case om.OS_Empty:
				if anyRemoving(ids, f, b) {
					// do nothing
				} else if anyCompleted(ids, f, b) {
					setStatus(om.OS_Completed, f, b)
				} else if anyExisting(ids, f, b) {
					if owner := getOwner(ids, f, b); owner >= 0 {
						addOrderWithOwner(f, b, owner)
					} else {
						addOrder(f, b)
					}
				}
			case om.OS_Existing:
				if anyRemoving(ids, f, b) {
					setStatus(om.OS_Removing, f, b)
				} else if anyCompleted(ids, f, b) {
					setStatus(om.OS_Completed, f, b)
				} else if anyExisting(ids, f, b) {
					if owner := getOwner(ids, f, b); owner >= 0 {
						addOrderWithOwner(f, b, owner)
					} else if _, bestId := getLowestCost(ids, f, b); allExisting(ids, f, b) && bestId == def.LocalID {
						takeOrder(f, b)
					}
				}
			case om.OS_Completed:
				if !anyExisting(ids, f, b) && !anyEmpty(ids, f, b) {
					setStatus(om.OS_Removing, f, b)
				}
			case om.OS_Removing:
				if !anyExisting(ids, f, b) && !anyCompleted(ids, f, b) {
					setStatus(om.OS_Empty, f, b)
				}
			}
		}
	}

}

func anyRemoving(ids []int, floor int, button driver.ButtonType) bool {
	for _, id := range ids {
		if om.GetOrders(id).GetStatus(floor, button) == om.OS_Removing {
			return true
		}
	}
	return false
}

func anyCompleted(ids []int, floor int, button driver.ButtonType) bool {
	for _, id := range ids {
		if om.GetOrders(id).GetStatus(floor, button) == om.OS_Completed {
			return true
		}
	}
	return false
}

func anyExisting(ids []int, floor int, button driver.ButtonType) bool {
	for _, id := range ids {
		if om.GetOrders(id).GetStatus(floor, button) == om.OS_Existing {
			return true
		}
	}
	return false
}

func allExisting(ids []int, floor int, button driver.ButtonType) bool {
	for _, id := range ids {
		if om.GetOrders(id).GetStatus(floor, button) != om.OS_Existing {
			return false
		}
	}
	return true
}

func anyEmpty(ids []int, floor int, button driver.ButtonType) bool {
	for _, id := range ids {
		if om.GetOrders(id).GetStatus(floor, button) == om.OS_Empty {
			return true
		}
	}
	return false
}

func setStatus(orderStatus om.OrderStatus, floor int, button driver.ButtonType) {
	if orderStatus == om.OS_Completed {
		om.GetOrders(def.LocalID).UpdateOrder(floor, button)
	} else if orderStatus == om.OS_Empty {
		om.GetOrders(def.LocalID).RemoveOrder(floor, button)
	} else {
		om.GetOrders(def.LocalID).SetStatus(floor, button, orderStatus)
		fsm.SetAllLights()
	}
}

func getOwner(ids []int, floor int, button driver.ButtonType) int {
	for _, id := range ids {
		if om.GetOrders(id).GetOwner(floor, button) >= 0 {
			return om.GetOrders(id).GetOwner(floor, button)
		}
	}
	return -1
}

func addOrderWithOwner(floor int, button driver.ButtonType, owner int) {
	scheduler.AddOrderWithOwner(fsm.Elevator, floor, button, owner)
	fsm.SetAllLights()
}

func addOrder(floor int, button driver.ButtonType) {
	scheduler.AddOrder(fsm.Elevator, floor, button)
}

func takeOrder(floor int, button driver.ButtonType) {
	om.GetOrders(def.LocalID).SetOwner(floor, button, def.LocalID)
	fsm.SetAllLights()
	fsm.OnNewOrder(floor, button)
}

// getLowestCost gets the lowest cost of the available ids and the corresponding id, or -1 if no elevator has registered its cost
func getLowestCost(ids []int, floor int, button driver.ButtonType) (cost int, id int) {
	bestId := ids[0]
	lowestCost := om.GetOrders(ids[0]).GetCost(floor, button)
	for _, id := range ids {
		if cost := om.GetOrders(id).GetCost(floor, button); lowestCost < 0 {
			lowestCost = cost
			bestId = id
		} else if cost >= 0 && cost < lowestCost {
			lowestCost = cost
			bestId = id
		}
	}
	return lowestCost, bestId
}
