package network

import (
	"fmt"
	"strconv"

	def "elevatorproject/definitions"
	"elevatorproject/network/bcast"
	"elevatorproject/network/peers"
	"elevatorproject/ordermanager"
)

type ordersMsg struct {
	ID     int
	Orders def.Matrix
}

func Init() {
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, strconv.Itoa(def.LocalID), peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	ordersTx := make(chan ordersMsg)
	ordersRx := make(chan ordersMsg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, ordersTx)
	go bcast.Receiver(16569, ordersRx)

	// get regular updates on the orders
	ordersUpdate := make(chan def.Matrix)
	go ordermanager.PollOrders(ordersUpdate)

	go func() {
		for {
			select {
			case p := <-peerUpdateCh:
				fmt.Printf("Peer update:\n")
				fmt.Printf("  Peers:    %q\n", p.Peers)
				fmt.Printf("  New:      %q\n", p.New)
				fmt.Printf("  Lost:     %q\n", p.Lost)

			case msg := <-ordersRx:
				if msg.ID != def.LocalID {
					ordermanager.AddMatrix(msg.ID, msg.Orders)
				}
			case orders := <-ordersUpdate:
				ordersTx <- ordersMsg{def.LocalID, orders}
			}
		}
	}()
}
