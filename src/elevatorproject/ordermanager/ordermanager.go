package ordermanager

import (
	"bytes"
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"fmt"
	"sync"
)

type OrderStatus int

const (
	OS_Empty     OrderStatus = 0
	OS_Existing  OrderStatus = 1
	OS_Completed OrderStatus = 2
	OS_Removing  OrderStatus = 3
)

type order struct {
	Status OrderStatus
	Cost   int
	Owner  int
}

var mutex = sync.Mutex{}
var orderMatrices [def.ElevatorCount]OrderMatrix

type OrderMatrix [def.FloorCount][def.ButtonCount]order

func init() {
	// Create empty order matrices
	m := OrderMatrix{}
	for f := 0; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			m[f][b] = createEmptyOrder()
		}
	}
	for elevId := 0; elevId < def.ElevatorCount; elevId++ {
		orderMatrices[elevId] = m
	}

}

func Lock() {
	mutex.Lock()
}

func Unlock() {
	mutex.Unlock()
}

func GetMatrix(id int) *OrderMatrix {
	return &orderMatrices[id]
}

func ButtonPressed(floor int, button driver.ButtonType) bool {
	return orderMatrices[def.LocalID].GetStatus(floor, button) != 0
}

func (m *OrderMatrix) GetStatus(floor int, button driver.ButtonType) OrderStatus {
	return m[floor][button].Status
}

func (m *OrderMatrix) SetStatus(floor int, button driver.ButtonType, status OrderStatus) {
	m[floor][button].Status = status
}

func (m *OrderMatrix) GetOwner(floor int, button driver.ButtonType) int {
	return m[floor][button].Owner
}

func (m *OrderMatrix) SetOwner(floor int, button driver.ButtonType, owner int) {
	m[floor][button].Owner = owner
}

func (m *OrderMatrix) GetCost(floor int, button driver.ButtonType) int {
	return m[floor][button].Cost
}

func (m *OrderMatrix) SetCost(floor int, button driver.ButtonType, cost int) {
	m[floor][button].Cost = cost
}

func (m *OrderMatrix) IsEmpty(floor int, button driver.ButtonType) bool {
	return m[floor][button].Status == OS_Empty
}

func (m *OrderMatrix) HasOrder(floor int, button driver.ButtonType) bool {
	return m[floor][button].Owner == def.LocalID && m[floor][button].Status == 1
}

func (m *OrderMatrix) HasSystemOrder(floor int, button driver.ButtonType) bool {
	return m[floor][button].Owner != -1 && m[floor][button].Status == 1
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
	if m[floor][button].Status == 1 {
		if button == driver.BT_Cab {
			m[floor][button] = createEmptyOrder()
		} else {
			m[floor][button].Status = 2
		}
	}
}

func (m *OrderMatrix) AddOrder(floor int, button driver.ButtonType, cost int) {
	if m[floor][button].Status == 0 {
		m[floor][button].Status = 1
		m[floor][button].Cost = cost
	}
}

func (m *OrderMatrix) AddCabOrder(floor int, owner int) {
	if m[floor][driver.BT_Cab].Status == 0 {
		m[floor][driver.BT_Cab].Status = 1
		m[floor][driver.BT_Cab].Owner = owner
	}
}

func AddMatrix(id int, newMatrix OrderMatrix) {
	orderMatrices[id] = newMatrix
}

func createEmptyOrder() order {
	return order{0, -1, -1}
}

func PrintOrder(orders OrderMatrix) {
	var buffer bytes.Buffer
	for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
		for f := 0; f < def.FloorCount; f++ {
			buffer.WriteString(fmt.Sprintf("%v:%v:%v", orders.GetStatus(f, b), orders.GetOwner(f, b), orders.GetCost(f, b)))
			if f != def.FloorCount-1 {
				buffer.WriteString(" | ")
			}
		}
		buffer.WriteString("\n")
	}
	fmt.Println(buffer.String())
}
