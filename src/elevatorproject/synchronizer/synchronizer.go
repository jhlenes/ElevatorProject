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

			// Reassign orders owned by <id> to the elevator with the lowest cost
			if om.GetMatrix(def.LocalID).GetOwner(f, b) == id {
				om.GetMatrix(def.LocalID).RemoveOrder(f, b)
				if cost, bestId := getCost(ids, f, b); cost >= 0 { // TODO: what if cost is < 0 ?
					setOwner(bestId, f, b)
					if bestId == def.LocalID {
						go fsm.OnNewOrder(f, b)
					}
				}
			}

			// TODO: Should also take order in some cases if communication was lost during confirmation
		}
	}
}

func Synchronize(ids []int) {
	for f := 0; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			if b == driver.BT_Cab {
				continue
			}

			switch om.GetMatrix(def.LocalID).GetStatus(f, b) {
			case om.OS_Empty:
				if anyCompleted(ids, f, b) {
					setStatus(om.OS_Completed, f, b)
				} else if anyExisting(ids, f, b) {
					if owner := getOwner(ids, f, b); owner >= 0 {
						copyOrder(owner, f, b)
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
						copyOrder(owner, f, b)
					} else if lowestCost, bestId := getCost(ids, f, b); lowestCost >= 0 && bestId == def.LocalID {
						setOwner(def.LocalID, f, b)
						fsm.OnNewOrder(f, b)
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
		if om.GetMatrix(id).GetStatus(floor, button) == om.OS_Removing {
			return true
		}
	}
	return false
}

func anyCompleted(ids []int, floor int, button driver.ButtonType) bool {
	for _, id := range ids {
		if om.GetMatrix(id).GetStatus(floor, button) == om.OS_Completed {
			return true
		}
	}
	return false
}

func anyExisting(ids []int, floor int, button driver.ButtonType) bool {
	for _, id := range ids {
		if om.GetMatrix(id).GetStatus(floor, button) == om.OS_Existing {
			return true
		}
	}
	return false
}

func anyEmpty(ids []int, floor int, button driver.ButtonType) bool {
	for _, id := range ids {
		if om.GetMatrix(id).GetStatus(floor, button) == om.OS_Empty {
			return true
		}
	}
	return false
}

func setStatus(orderStatus om.OrderStatus, floor int, button driver.ButtonType) {
	if orderStatus == om.OS_Completed {
		om.GetMatrix(def.LocalID).UpdateOrder(floor, button)
	} else if orderStatus == om.OS_Empty {
		om.GetMatrix(def.LocalID).RemoveOrder(floor, button)
	} else {
		om.GetMatrix(def.LocalID).SetStatus(floor, button, orderStatus)
		fsm.SetAllLights()
	}
}

func getOwner(ids []int, floor int, button driver.ButtonType) int {
	for _, id := range ids {
		if om.GetMatrix(id).GetOwner(floor, button) >= 0 {
			return om.GetMatrix(id).GetOwner(floor, button)
		}
	}
	return -1
}

func copyOrder(owner int, floor int, button driver.ButtonType) {
	om.GetMatrix(def.LocalID).SetStatus(floor, button, om.OS_Existing)
	om.GetMatrix(def.LocalID).SetOwner(floor, button, owner)
	fsm.SetAllLights()
}

func addOrder(floor int, button driver.ButtonType) {
	scheduler.AddOrder(fsm.Elevator, floor, button)
}

func setOwner(id int, f int, b driver.ButtonType) {
	om.GetMatrix(def.LocalID).SetOwner(f, b, id)
	fsm.SetAllLights()
}

func getCost(ids []int, f int, b driver.ButtonType) (cost int, id int) {
	bestId := ids[0]
	lowestCost := om.GetMatrix(ids[0]).GetCost(f, b)
	for _, id := range ids {
		if om.GetMatrix(id).GetCost(f, b) < lowestCost {
			lowestCost = om.GetMatrix(id).GetCost(f, b)
			bestId = id
		}
	}
	return lowestCost, bestId
}
