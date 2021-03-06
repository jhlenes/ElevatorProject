package ordermanager

import (
	"bytes"
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"encoding/json"
	"fmt"
	"os"
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
	return orderMatrices[def.LocalId][floor][button].Status != 0
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
	return m[floor][button].Owner == def.LocalId && m[floor][button].Status == 1
}

func (m *OrderMatrix) HasOrderOnFloor(floor int) bool {
	for b := driver.ButtonType(0); b < def.ButtonCount; b++ {
		if m[floor][b].Owner == def.LocalId && m[floor][b].Status == 1 {
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
	if button == driver.BT_Cab {
		backupCabOrders()
	}
}

func (m *OrderMatrix) AddOrder(floor int, button driver.ButtonType, cost int) {
	m[floor][button].Status = 1
	m[floor][button].Cost = cost
}

func (m *OrderMatrix) AddCabOrder(floor int, owner int) {
	m[floor][driver.BT_Cab].Status = 1
	m[floor][driver.BT_Cab].Owner = owner
	backupCabOrders()
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
			buffer.WriteString(fmt.Sprintf("(%v,%v,%v)", orders[f][b].Status, orders[f][b].Cost, orders[f][b].Owner))
			if f != def.FloorCount-1 {
				buffer.WriteString(" | ")
			}
		}
		buffer.WriteString("\n")
	}
	fmt.Println(buffer.String())
}

func backupCabOrders() {
	fileName := fmt.Sprintf("backup_%v.dat", def.LocalId)
	os.Remove(fileName)
	f, err1 := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0755)
	if err1 != nil {
		def.Info.Println("Could not make backup.")
		return
	}
	b, err2 := json.Marshal(orderMatrices[def.LocalId])
	if err2 != nil {
		def.Info.Println("Could not make backup.")
		return
	}
	f.Write(b)
	f.Close()
}

func ReadBackup() {
	fileName := fmt.Sprintf("backup_%v.dat", def.LocalId)
	f, err1 := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0755)
	defer f.Close()
	if err1 != nil {
		def.Info.Println("No backup.")
		return
	}
	b := make([]byte, 2048)
	n, err2 := f.Read(b)
	if err2 != nil {
		def.Info.Println("No backup.")
		return
	}
	m := OrderMatrix{}
	err3 := json.Unmarshal(b[:n], &m)
	if err3 != nil {
		def.Info.Println("No backup.")
		return
	}

	def.Info.Println("Backup loaded!")

	for f := 0; f < def.FloorCount; f++ {
		if m[f][driver.BT_Cab].Status == 1 {
			orderMatrices[def.LocalId][f][driver.BT_Cab] = m[f][driver.BT_Cab]
		}
	}
}
