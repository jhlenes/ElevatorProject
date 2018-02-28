package synchronizer

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/fsm"
	om "elevatorproject/ordermanager"
	"elevatorproject/scheduler"
	"log"
	"strconv"
)

func getIdsFromPeers(peers []string) []int {
	ids := []int{}
	for _, peer := range peers {
		id, err := strconv.Atoi(peer)
		if err != nil {
			log.Println("Id of a peer is not an integer.")
		} else {
			ids = append(ids, id)
		}
	}
	return ids
}

func ReassignOrders(peers []string, id int) {
	ids := getIdsFromPeers(peers)

	def.Info.Printf("Reassigning orders of %v to elevators: %v\n", id, ids)

	for f := 0; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {

			// Reassign orders owned by <id> to the elevator with the lowest cost
			if om.OrderMatrices[def.LocalID][f][b].Owner == id {
				om.OrderMatrices[id].RemoveOrder(f, b)
				if cost, bestId := getCost(ids, f, b); cost >= 0 { // TODO: what if cost is < 0 ?
					setOwner(bestId, f, b)
					if bestId == def.LocalID {
						fsm.OnNewOrder(f, b)
					}
				}
			}

			// TODO: Should also take order in some cases if communication was lost during confirmation
		}
	}
}

func Synchronize(peers []string, new string, lost []string) {
	ids := getIdsFromPeers(peers)

	for f := 0; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			if b == driver.BT_Cab {
				continue
			}

			switch om.OrderMatrices[def.LocalID][f][b].Status {
			case om.OS_Empty:
				if anyRemoving(ids, f, b) {
					// do nothing
				} else if anyCompleted(ids, f, b) {
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
		if om.OrderMatrices[id][floor][button].Status == om.OS_Removing {
			return true
		}
	}
	return false
}

func anyCompleted(ids []int, floor int, button driver.ButtonType) bool {
	for _, id := range ids {
		if om.OrderMatrices[id][floor][button].Status == om.OS_Completed {
			return true
		}
	}
	return false
}

func anyExisting(ids []int, floor int, button driver.ButtonType) bool {
	for _, id := range ids {
		if om.OrderMatrices[id][floor][button].Status == om.OS_Existing {
			return true
		}
	}
	return false
}

func anyEmpty(ids []int, floor int, button driver.ButtonType) bool {
	for _, id := range ids {
		if om.OrderMatrices[id][floor][button].Status == om.OS_Empty {
			return true
		}
	}
	return false
}

func setStatus(orderStatus om.OrderStatus, floor int, button driver.ButtonType) {
	if orderStatus == om.OS_Completed {
		om.GetLocalOrderMatrix().UpdateOrder(floor, button)
	} else if orderStatus == om.OS_Empty {
		om.GetLocalOrderMatrix().RemoveOrder(floor, button)
	} else {
		om.OrderMatrices[def.LocalID][floor][button].Status = orderStatus
		fsm.SetAllLights()
	}
}

func getOwner(ids []int, floor int, button driver.ButtonType) int {
	for _, id := range ids {
		if om.OrderMatrices[id][floor][button].Owner >= 0 {
			return om.OrderMatrices[id][floor][button].Owner
		}
	}
	return -1
}

func copyOrder(owner int, floor int, button driver.ButtonType) {
	om.OrderMatrices[def.LocalID][floor][button].Status = om.OS_Existing
	om.OrderMatrices[def.LocalID][floor][button].Owner = owner
	fsm.SetAllLights()
}

func addOrder(floor int, button driver.ButtonType) {
	scheduler.AddOrder(fsm.Elevator, floor, button)
}

func setOwner(id int, f int, b driver.ButtonType) {
	om.OrderMatrices[def.LocalID][f][b].Owner = id
	fsm.SetAllLights()
}

func getCost(ids []int, f int, b driver.ButtonType) (cost int, id int) {
	bestId := ids[0]
	lowestCost := om.OrderMatrices[ids[0]][f][b].Cost
	for _, id := range ids {
		if om.OrderMatrices[id][f][b].Cost < lowestCost {
			lowestCost = om.OrderMatrices[id][f][b].Cost
			bestId = id
		}
	}
	return lowestCost, bestId
}
