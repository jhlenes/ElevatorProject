package network

import (
	"fmt"
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

func Init() {

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
	ordersUpdate := make(chan ordermanager.OrderMatrix, 10)
	go func() {
		for {
			time.Sleep(def.SendTime * time.Millisecond)
			ordersUpdate <- *ordermanager.GetLocalOrderMatrix()
		}
	}()

	elevators := peers.PeerUpdate{}
	go func() {
		for {
			select {
			case p := <-peerUpdateCh:
				fmt.Printf("Peer update:\n")
				fmt.Printf("  Peers:    %q\n", p.Peers)
				fmt.Printf("  New:      %q\n", p.New)
				fmt.Printf("  Lost:     %q\n", p.Lost)
				elevators = p

			case msg := <-ordersRx:
				if msg.ID != def.LocalID {
					ordermanager.AddMatrix(msg.ID, msg.Orders)
					synchronizer.Synchronize(elevators.Peers, elevators.New, elevators.Lost)
				}
			case orders := <-ordersUpdate:
				ordersTx <- ordersMsg{def.LocalID, orders}
			}
		}
	}()
}
