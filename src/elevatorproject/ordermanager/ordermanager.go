package ordermanager

import (
	def "elevatorproject/definitions"
)

func GetOrder(floor int, button def.ButtonType) bool {
	return false
}

func OrdersAbove(floor int) bool {
	return true
}

func OrdersBelow(floor int) bool {
	return true
}
