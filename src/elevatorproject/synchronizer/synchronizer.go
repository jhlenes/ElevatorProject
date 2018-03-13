package synchronizer

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/fsm"
	om "elevatorproject/ordermanager"
	"elevatorproject/scheduler"
)

func StartOperatingAlone() {
	for f := 0; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			if status := om.GetOrders(def.LocalId).GetStatus(f, b); status == om.OS_Completed || status == om.OS_Removing {
				om.GetOrders(def.LocalId).RemoveOrder(f, b)
			}
		}
	}
}

func ReassignOrders(ids []int, id int) {
	if len(ids) == 0 {
		return
	}
	def.Info.Printf("Reassigning orders of %v to elevators: %v\n", id, ids)

	for f := 0; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			if b == driver.BT_Cab {
				continue
			}

			// Reassign orders owned by <id> to the elevator with the lowest cost
			if om.GetOrders(def.LocalId).GetOwner(f, b) == id {
				if cost, bestId := getLowestCost(ids, f, b); cost >= 0 {
					addOrderWithOwner(f, b, bestId)
				} else { // couldn't find a new owner
					addOrder(f, b) // TODO: correct?
				}
			}

			// TODO: Should also take order in some cases if communication was lost during confirmation
			if om.GetOrders(def.LocalId).GetStatus(f, b) == om.OS_Existing && om.GetOrders(id).GetStatus(f, b) == om.OS_Empty {
				if om.GetOrders(def.LocalId).GetOwner(f, b) < 0 && om.GetOrders(id).GetOwner(f, b) < 0 {
					if cost, bestId := getLowestCost(ids, f, b); cost >= 0 {
						addOrderWithOwner(f, b, bestId)
					} else {
						takeOrder(f, b)
					}
				}
			}

		}
	}
}

func Synchronize(onlineIds, activeIds []int) {
	if len(onlineIds) == 0 {
		return
	}
	for f := 0; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			if b == driver.BT_Cab {
				continue
			}

			switch om.GetOrders(def.LocalId).GetStatus(f, b) {
			case om.OS_Empty:
				if anyRemoving(onlineIds, f, b) {
					// do nothing
				} else if anyCompleted(onlineIds, f, b) {
					setStatus(om.OS_Completed, f, b)
				} else if anyExisting(onlineIds, f, b) {
					if owner := getOwner(onlineIds, f, b); owner >= 0 {
						addOrderWithOwner(f, b, owner)
					} else {
						addOrder(f, b)
					}
				}
			case om.OS_Existing:
				if anyRemoving(onlineIds, f, b) {
					setStatus(om.OS_Removing, f, b)
				} else if anyCompleted(onlineIds, f, b) {
					setStatus(om.OS_Completed, f, b)
				} else if anyExisting(onlineIds, f, b) && om.GetOrders(def.LocalId).GetOwner(f, b) < 0 {
					if owner := getOwner(onlineIds, f, b); owner >= 0 {
						setOwner(f, b, owner)
					} else if shouldTakeOrder(f, b, onlineIds, activeIds) {
						takeOrder(f, b)
					}
				}
			case om.OS_Completed:
				if !anyExisting(onlineIds, f, b) && !anyEmpty(onlineIds, f, b) {
					setStatus(om.OS_Removing, f, b)
				}
			case om.OS_Removing:
				if !anyExisting(onlineIds, f, b) && !anyCompleted(onlineIds, f, b) {
					setStatus(om.OS_Empty, f, b)
				}
			}
		}
	}
}

func shouldTakeOrder(floor int, button driver.ButtonType, onlineIds, activeIds []int) bool {
	lowestCost, bestId := getLowestCost(onlineIds, floor, button)
	return allExisting(onlineIds, floor, button) && bestId == def.LocalId && lowestCost >= 0 && len(activeIds) >= 2
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
	if orderStatus == om.OS_Empty {
		om.GetOrders(def.LocalId).RemoveOrder(floor, button)
	} else {
		om.GetOrders(def.LocalId).SetStatus(floor, button, orderStatus)
		fsm.SetAllLights()
	}
}

// getOwner returns the owner of an order if one exist, else -1
func getOwner(ids []int, floor int, button driver.ButtonType) int {
	for _, id := range ids {
		if om.GetOrders(id).GetOwner(floor, button) >= 0 {
			return om.GetOrders(id).GetOwner(floor, button)
		}
	}
	return -1
}

func setOwner(floor int, button driver.ButtonType, owner int) {
	om.GetOrders(def.LocalId).SetOwner(floor, button, owner)
	fsm.SetAllLights()
	if owner == def.LocalId {
		fsm.OnNewOrder(floor, button)
	}
}

func addOrderWithOwner(floor int, button driver.ButtonType, owner int) {
	scheduler.AddOrderWithOwner(fsm.Elevator, floor, button, owner)
	fsm.SetAllLights()
	if owner == def.LocalId {
		fsm.OnNewOrder(floor, button)
	}
}

func addOrder(floor int, button driver.ButtonType) {
	scheduler.AddOrder(fsm.Elevator, floor, button)
}

func takeOrder(floor int, button driver.ButtonType) {
	om.GetOrders(def.LocalId).SetOwner(floor, button, def.LocalId)
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
