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

func Synchronize(peers []string, new string, lost []string) {
	//ids := getIdsFromPeers(peers)

	id := 0
	if def.LocalID == 0 {
		id = 1
	}

	for f := 0; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			if b == driver.BT_Cab {
				continue
			}


			// assuming 2 elevators
			if om.OrderMatrices[def.LocalID][f][b].Status < om.OrderMatrices[id][f][b].Status { // update status to higher number

				if om.OrderMatrices[id][f][b].Status == om.OS_Existing { // new order
					scheduler.AddOrder(fsm.Elevator, f, b)
					
				} else if om.OrderMatrices[id][f][b].Status == om.OS_Completed { // finished order
					om.OrderMatrices[def.LocalID][f][b].Status = om.OS_Completed
				} else if om.OrderMatrices[def.LocalID][f][b].Status != om.OS_Empty && om.OrderMatrices[id][f][b].Status == om.OS_Removing { // deleting order
					om.OrderMatrices[def.LocalID][f][b].Status = om.OS_Removing
				}
				fsm.SetAllLights()

			} else if om.OrderMatrices[def.LocalID][f][b].Status == om.OrderMatrices[id][f][b].Status { // equal status

				if om.OrderMatrices[id][f][b].Status == om.OS_Existing && om.OrderMatrices[def.LocalID][f][b].Owner < 0 { // both status 1 => decide owner
					if om.OrderMatrices[id][f][b].Owner >= 0 { // copy owner
						om.OrderMatrices[def.LocalID][f][b].Owner = om.OrderMatrices[id][f][b].Owner
					} else { // decide owner
						if om.OrderMatrices[def.LocalID][f][b].Cost < om.OrderMatrices[id][f][b].Cost {
							om.OrderMatrices[def.LocalID][f][b].Owner = def.LocalID
						} else {
							om.OrderMatrices[def.LocalID][f][b].Owner = id
						}
					}
					if om.OrderMatrices[def.LocalID][f][b].Owner == def.LocalID {
						fsm.OnNewOrder(f, b)
						def.Info.Print("We have a new order")
					}
				} else if om.OrderMatrices[id][f][b].Status == om.OS_Completed { //  TODO: check that all have finished order
					om.OrderMatrices[def.LocalID][f][b].Status = om.OS_Removing
				} else if om.OrderMatrices[def.LocalID][f][b].Status == om.OS_Removing {
					om.OrderMatrices[def.LocalID].RemoveOrder(f, b)
					fsm.SetAllLights()
				}
			} else if om.OrderMatrices[def.LocalID][f][b].Status == om.OS_Removing && om.OrderMatrices[id][f][b].Status == om.OS_Empty { //
				om.OrderMatrices[def.LocalID].RemoveOrder(f, b)
				fsm.SetAllLights()
			}
		}
	}


}
