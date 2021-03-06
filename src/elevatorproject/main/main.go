package main

import (
	def "elevatorproject/definitions"
	"elevatorproject/driver"
	"elevatorproject/fsm"
	"elevatorproject/network"
	om "elevatorproject/ordermanager"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func main() {

	// Get the id, address and port from command line arguments, or use defaults
	var id int
	var addr string
	var port int
	flag.IntVar(&id, "id", def.LocalId, "id of this peer")
	flag.StringVar(&addr, "addr", def.Addr, "address of elevator server")
	flag.IntVar(&port, "port", def.Port, "port of elevator server")
	flag.Parse()
	def.LocalId = id
	def.Addr = addr
	def.Port = port

	fsm.Init()
	network.Init()

	go printGoroutineStackTracesOnSigquit()
	waitForShutdownSignal()
}

// waitForShutdownSignal shutdowns the program in a safe way when terminated by user (ctrl+c)
func waitForShutdownSignal() {
	var c = make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	<-c
	driver.SetMotorDirection(driver.MD_Stop)
	def.Info.Println("User terminated the program.")
	os.Exit(1)
}

// printGoroutineStackTracesOnSigquit is called when you press ^\ (Control+Backslash) and prints the goroutine stacktraces,
// the order matrix and the elevator states
func printGoroutineStackTracesOnSigquit() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGQUIT)
	buf := make([]byte, 1<<20)
	for {
		<-sigs
		stacklen := runtime.Stack(buf, true)
		log.Printf("=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end\n", buf[:stacklen])
		om.PrintOrder(*om.GetOrders(def.LocalId).(*om.OrderMatrix))
		fmt.Printf("%+v\n", fsm.Elevator)
	}
}