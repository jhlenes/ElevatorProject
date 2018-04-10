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
	Stuck  bool
	Orders ordermanager.OrderMatrix
}

var elevatorTimers [def.ElevatorCount]*time.Timer
var onlineElevators = make(map[int]bool)
var activeElevators = make(map[int]bool)

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
	for {
		select {
		case msg := <-ordersRx:
			if fsm.Elevator.Behaviour != def.Initializing {
				checkForNewOrStuckElevators(msg)

				elevatorTimers[msg.ID].Reset(def.ElevatorTimeout * time.Second)
				if msg.ID != def.LocalId {
					ordermanager.AddMatrix(msg.ID, msg.Orders)
					synchronizer.Synchronize(getIds(onlineElevators), getIds(activeElevators))
				}
			}
		case lostId := <-elevatorTimeoutCh:
			if lostId != def.LocalId {
				delete(onlineElevators, lostId)
				delete(activeElevators, lostId)
				fsm.NumActiveElevators = len(activeElevators)
				def.Info.Printf("Peers: %v\n", getIds(onlineElevators))
				elevatorTimers[lostId].Stop()

				if len(onlineElevators) < 2 {
					synchronizer.StartOperatingAlone()
				}

				if fsm.Elevator.Behaviour != def.Initializing {
					synchronizer.ReassignOrders(getIds(activeElevators), lostId)
				}
			}
		}
	}
}

func sendMessages(ordersTx chan ordersMsg) {
	for {
		time.Sleep(def.SendTime * time.Millisecond)
		if len(onlineElevators) < 2 { // TODO: maybe move this elesewhere?
			synchronizer.StartOperatingAlone()
		}
		isStuck := fsm.Elevator.Stuck
		ordersTx <- ordersMsg{def.LocalId, isStuck, *ordermanager.GetOrders(def.LocalId).(*ordermanager.OrderMatrix)}

	}
}

func getIds(elevatorMap map[int]bool) []int {
	i := 0
	ids := make([]int, len(elevatorMap))
	for id := range elevatorMap {
		ids[i] = id
		i++
	}
	return ids
}

func checkForNewOrStuckElevators(msg ordersMsg) {
	if _, ok := onlineElevators[msg.ID]; !ok { // if new id
		onlineElevators[msg.ID] = true
		activeElevators[msg.ID] = true
		fsm.NumActiveElevators = len(activeElevators)
		def.Info.Printf("Peers: %v\n", getIds(onlineElevators))
	} else if _, ok := activeElevators[msg.ID]; msg.Stuck && ok { // elevator is stuck
		if msg.ID != def.LocalId {
			delete(activeElevators, msg.ID)
			fsm.NumActiveElevators = len(activeElevators)
			elevatorTimers[msg.ID].Stop()
			def.Info.Printf("Elevator is stuck: %v\n", msg.ID)
			synchronizer.ReassignOrders(getIds(activeElevators), msg.ID)
		}
	} else if !msg.Stuck && !ok { // elevator is no longer stuck
		activeElevators[msg.ID] = true
		fsm.NumActiveElevators = len(activeElevators)
		def.Info.Printf("Elevator no longer stuck: %v\n", msg.ID)
	}
}
