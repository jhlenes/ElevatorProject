package network

import (
	"strconv"
	"time"

	def "elevatorproject/definitions"
	"elevatorproject/network/bcast"
	"elevatorproject/network/peers"
	"elevatorproject/ordermanager"
	"elevatorproject/synchronizer"
)

type ordersMsg struct {
	ID     int
	Orders ordermanager.OrderMatrix
}

var elevatorTimers [def.ElevatorCount]*time.Timer

func Init() {

	elevatorTimeoutCh := make(chan int, 10)

	// Create timers for each elevator and have them send the elevator id to a shared channel on timeout
	for i := range elevatorTimers {
		elevatorTimers[i] = time.NewTimer(def.ElevatorTimeout * time.Second)
		elevatorTimers[i].Stop()
		go func(timer *time.Timer, id int) {
			for range timer.C {
				elevatorTimeoutCh <- id
			}
		}(elevatorTimers[i], i)
	}

	// Listen for other peers and send own status
	peerUpdateCh := make(chan peers.PeerUpdate, 10)
	peerTxEnable := make(chan bool, 10)
	go peers.Transmitter(15699, strconv.Itoa(def.LocalID), peerTxEnable)
	go peers.Receiver(15699, peerUpdateCh)

	// Set up channels for broadcasting and listening for orders
	ordersTx := make(chan ordersMsg, 10)
	ordersRx := make(chan ordersMsg, 10)
	go bcast.Transmitter(16539, ordersTx)
	go bcast.Receiver(16539, ordersRx)

	// send regular updates on the orders
	ordersUpdate := make(chan ordermanager.OrderMatrix, 1000)
	go func() {
		for {
			time.Sleep(def.SendTime * time.Millisecond)
			ordersUpdate <- *ordermanager.GetMatrix(def.LocalID)
		}
	}()

	onlineElevators := make(map[int]bool)
	go func() {
		for {
			select {
			/*case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)*/

			case msg := <-ordersRx:
				if _, ok := onlineElevators[msg.ID]; !ok {
					onlineElevators[msg.ID] = true
					def.Info.Printf("Peers: %v\n", getKeys(onlineElevators))
				}
				if msg.ID != def.LocalID {
					elevatorTimers[msg.ID].Reset(def.ElevatorTimeout * time.Second)
					if msg.Orders != *ordermanager.GetMatrix(msg.ID) {
						ordermanager.AddMatrix(msg.ID, msg.Orders)
					}
					go synchronizer.Synchronize(getKeys(onlineElevators))
				}

			case orders := <-ordersUpdate:
				ordersTx <- ordersMsg{def.LocalID, orders}

			case id := <-elevatorTimeoutCh:
				delete(onlineElevators, id)
				def.Info.Printf("Peers: %v\n", getKeys(onlineElevators))
				elevatorTimers[id].Stop()
				go synchronizer.ReassignOrders(getKeys(onlineElevators), id)
			}
		}
	}()
}

func getKeys(mymap map[int]bool) []int {
	i := 0
	keys := make([]int, len(mymap))
	for k := range mymap {
		keys[i] = k
		i++
	}
	return keys
}
