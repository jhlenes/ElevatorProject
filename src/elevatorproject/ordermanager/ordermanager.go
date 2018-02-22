package ordermanager

import (
	"bytes"
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"fmt"
)

type order struct {
	Status int
	Cost   int
	Owner  int
}

type OrderMatrix [def.FloorCount][def.ButtonCount]order

// TODO: maybe have a lock?
var OrderMatrices [def.ElevatorCount]OrderMatrix

func init() {
	// Create empty order matrices
	m := OrderMatrix{}
	for f := 0; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			m[f][b] = createEmptyOrder()
		}
	}
	for elevId := 0; elevId < def.ElevatorCount; elevId++ {
		OrderMatrices[elevId] = m
	}
}

func GetLocalOrderMatrix() *OrderMatrix {
	return &OrderMatrices[def.LocalID]
}

func HasSystemOrder(floor int, button driver.ButtonType, ids []int) bool {
	for _, id := range ids {
		if OrderMatrices[id][floor][button].Status != 1 {
			return false
		}
	}
	return true
}

func (m *OrderMatrix) HasOrder(floor int, button driver.ButtonType) bool {
	return m[floor][button].Owner == def.LocalID && m[floor][button].Status == 1
}

func (m *OrderMatrix) HasOrderAbove(floor int) bool {
	for f := floor + 1; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			if m.HasOrder(f, b) {
				return true
			}
		}
	}
	return false
}

func (m *OrderMatrix) HasOrderBelow(floor int) bool {
	for f := 0; f < floor; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			if m.HasOrder(f, b) {
				return true
			}
		}
	}
	return false
}

func (m *OrderMatrix) RemoveOrder(floor int, button driver.ButtonType) {
	m[floor][button] = createEmptyOrder()
}

func (m *OrderMatrix) UpdateOrder(floor int, button driver.ButtonType) {
	m[floor][button].Status = 2
}

func (m *OrderMatrix) AddOrder(floor int, button driver.ButtonType, cost int) {
	m[floor][button].Status = 1
	m[floor][button].Cost = cost
}

func AddMatrix(id int, newMatrix OrderMatrix) {
	OrderMatrices[id] = newMatrix
}

func createEmptyOrder() order {
	return order{0, 9999, -1}
}

func printOrder(orders OrderMatrix) {
	var buffer bytes.Buffer
	for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
		for f := 0; f < def.FloorCount; f++ {
			buffer.WriteString(fmt.Sprintf("%v|%v|%v  ", orders[f][b].Status, orders[f][b].Cost, orders[f][b].Owner))
		}
		buffer.WriteString("\n")
	}
	fmt.Println(buffer.String())
}
