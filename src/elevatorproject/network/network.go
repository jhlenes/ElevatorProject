package network

import (
	"time"

	def "elevatorproject/definitions"
	"elevatorproject/fsm"
	"elevatorproject/network/bcast"
	"elevatorproject/ordermanager"
	"elevatorproject/synchronizer"
)

type ordersMsg struct {
	ID     int
	Orders ordermanager.OrderMatrix
}

var elevatorTimers [def.ElevatorCount]*time.Timer

func Init() {

	// Create timers for each elevator and have them send the elevator id to a shared channel on timeout
	elevatorTimeoutCh := make(chan int, 10)
	for i := range elevatorTimers {
		elevatorTimers[i] = time.NewTimer(def.ElevatorTimeout * time.Second)
		elevatorTimers[i].Stop()
		go func(timer *time.Timer, id int) {
			for range timer.C {
				elevatorTimeoutCh <- id
			}
		}(elevatorTimers[i], i)
	}

	// Set up channels for broadcasting and listening
	ordersTx := make(chan ordersMsg, 10)
	ordersRx := make(chan ordersMsg, 10)
	go bcast.Transmitter(16539, ordersTx)
	go bcast.Receiver(16539, ordersRx)

	go listenAtChannels(ordersRx, elevatorTimeoutCh)
	go sendMessages(ordersTx)
}

func listenAtChannels(ordersRx chan ordersMsg, elevatorTimeoutCh chan int) {
	onlineElevators := make(map[int]bool)
	for {
		select {
		case msg := <-ordersRx:
			if fsm.Elevator.Behaviour != def.Initializing {
				if _, ok := onlineElevators[msg.ID]; !ok { // if new id
					onlineElevators[msg.ID] = true
					def.Info.Printf("Peers: %v\n", getIds(onlineElevators))
				}
				elevatorTimers[msg.ID].Reset(def.ElevatorTimeout * time.Second)
				if msg.ID != def.LocalID {
					ordermanager.AddMatrix(msg.ID, msg.Orders)
					synchronizer.Synchronize(getIds(onlineElevators))
				}
			}
		case lostId := <-elevatorTimeoutCh:
			delete(onlineElevators, lostId)
			def.Info.Printf("Peers: %v\n", getIds(onlineElevators))
			elevatorTimers[lostId].Stop()
			if fsm.Elevator.Behaviour != def.Initializing && fsm.Elevator.Behaviour != def.Stuck {
				synchronizer.ReassignOrders(getIds(onlineElevators), lostId)
			}
		}
	}
}

func sendMessages(ordersTx chan ordersMsg) {
	for {
		time.Sleep(def.SendTime * time.Millisecond)
		if fsm.Elevator.Behaviour != def.Stuck {
			ordersTx <- ordersMsg{def.LocalID, *ordermanager.GetOrders(def.LocalID).(*ordermanager.OrderMatrix)}
		}
	}
}

func getIds(onlineElevators map[int]bool) []int {
	i := 0
	ids := make([]int, len(onlineElevators))
	for id := range onlineElevators {
		ids[i] = id
		i++
	}
	return ids
}
