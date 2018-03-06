package ordermanager

import (
	"bytes"
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"fmt"
)

type OrderStatus int

const (
	OS_Empty     OrderStatus = 0
	OS_Existing  OrderStatus = 1
	OS_Completed OrderStatus = 2
	OS_Removing  OrderStatus = 3
)

type Order struct {
	Status OrderStatus
	Cost   int
	Owner  int
}

type OrderMatrix [def.FloorCount][def.ButtonCount]Order

type Orders interface {
	GetStatus(floor int, button driver.ButtonType) OrderStatus
	SetStatus(floor int, button driver.ButtonType, status OrderStatus)
	GetOwner(floor int, button driver.ButtonType) int
	SetOwner(floor int, button driver.ButtonType, owner int)
	GetCost(floor int, button driver.ButtonType) int
	SetCost(floor int, button driver.ButtonType, cost int)
	GetOrder(floor int, button driver.ButtonType) Order
	SetOrder(floor int, button driver.ButtonType, order Order)

	HasOrder(floor int, button driver.ButtonType) bool
	HasOrderOnFloor(floor int) bool
	HasSystemOrder(floor int, button driver.ButtonType) bool
	HasOrderAbove(floor int) bool
	HasOrderBelow(floor int) bool

	RemoveOrder(floor int, button driver.ButtonType)
	UpdateOrder(floor int, button driver.ButtonType)
	AddOrder(floor int, button driver.ButtonType, cost int)
	AddCabOrder(floor int, owner int)
}

var orderMatrices [def.ElevatorCount]OrderMatrix

func init() {
	// Create empty order matrices
	m := OrderMatrix{}
	for f := 0; f < def.FloorCount; f++ {
		for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
			m[f][b] = CreateEmptyOrder()
		}
	}
	for elevId := 0; elevId < def.ElevatorCount; elevId++ {
		orderMatrices[elevId] = m
	}
}

func GetOrders(id int) Orders {
	return &orderMatrices[id]
}

func ButtonPressed(floor int, button driver.ButtonType) bool {
	return orderMatrices[def.LocalID][floor][button].Status != 0
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

func (m *OrderMatrix) GetOrder(floor int, button driver.ButtonType) Order {
	return m[floor][button]
}

func (m *OrderMatrix) SetOrder(floor int, button driver.ButtonType, order Order) {
	m[floor][button] = order
}

func (m *OrderMatrix) HasOrder(floor int, button driver.ButtonType) bool {
	return m[floor][button].Owner == def.LocalID && m[floor][button].Status == 1
}

func (m *OrderMatrix) HasOrderOnFloor(floor int) bool {
	for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
		if m[floor][b].Owner == def.LocalID && m[floor][b].Status == 1 {
			return true
		}
	}
	return false
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
	m[floor][button] = CreateEmptyOrder()
}

func (m *OrderMatrix) UpdateOrder(floor int, button driver.ButtonType) {
	if m[floor][button].Status == 1 {
		if button == driver.BT_Cab {
			m[floor][button] = CreateEmptyOrder()
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

func CreateEmptyOrder() Order {
	return Order{0, -1, -1}
}

func PrintOrder(orders OrderMatrix) {
	var buffer bytes.Buffer
	for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
		for f := 0; f < def.FloorCount; f++ {
			buffer.WriteString(fmt.Sprintf("%v", orders[f][b].Cost))
			if f != def.FloorCount-1 {
				buffer.WriteString(" | ")
			}
		}
		buffer.WriteString("\n")
	}
	fmt.Println(buffer.String())
}
