package main

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/fsm"
	"elevatorproject/network"
	"flag"
	"os"
	"os/signal"
)

func main() {

	// Get the id, address and port from command line arguments, or use defaults
	var id int
	var addr string
	var port int
	flag.IntVar(&id, "id", def.LocalID, "id of this peer")
	flag.StringVar(&addr, "addr", def.Addr, "address of elevator server")
	flag.IntVar(&port, "port", def.Port, "port of elevator server")
	flag.Parse()
	def.LocalID = id
	def.Addr = addr
	def.Port = port

	fsm.Init()
	network.Init()

	safeShutdown()
}

// safeShutdown shutdowns the program in a safe way when terminataed by user (ctrl+c)
func safeShutdown() {
	var c = make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
	driver.SetMotorDirection(driver.MD_Stop)
	def.Info.Println("User terminated the program.")
	os.Exit(1)
}
