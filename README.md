# Elevator project

This program implements a distributed fault tolerant elevator system in Go. The program has been developed for three elevators and four floors. However, it also works for an arbitrary amount of elevators and floors.

The key part of this project was making the system robust for all kinds of failures, i.e. an order should never be lost. E.g. If one elevator fails, another elevator should take its orders. The main idea behind this implementation is that the elevators constantly broadcasts their worldview, i.e. the placed orders, over UDP. All online elevators then have access to all the worldviews and can reach consensus without the need for acknowledgements.

This program received full score on the FAT test except for item 2.7 in the [specification](project-description/SPECIFICATION.md) which was misinterpreted. When only two orders existed in the system, one hall-up and one hall-down on the same floor, the elevator would clear both at the same time instead of waiting for some time to allow new cab orders to be placed. A quick fix for this was applied in the latest commit, but has not yet been tested extensively.

## Modules
This program consists of several modules:

### definitions
The definitions module contains program wide constants and types.

### driver
The driver module, found [here](https://github.com/TTK4145/driver-go), communicates with the hardware.

### fsm
The fsm module is an implementation of a finite state machine for the elevator. It sends commands to the driver, gets information about the orders through the scheduler, and also reads from the ordermanager in order to update the button lamps.

### main
The main module is the starting point of the program. It initializes the fsm and network modules, and waits for the shutdown signal.

### network
The network module is based on [this code](https://github.com/TTK4145/Network-go). It handles the communication between the elevators and forwards the received information to the synchronizer module. It also updates the fsm module with the number of active elevators.

### ordermanager
The ordermanager module stores the information about the orders.

### scheduler
The scheduler module uses information from the ordermanager to make decisions about what the elevator should do.

### synchronizer
The synchronizer module synchronizes the information received over the network with your own information stored in the ordermanager. It also handles reassignment of orders when elevators fail.


## Running on the actual hardware

Start the [elevator server](https://github.com/TTK4145/elevator-server)
by calling `ElevatorServer` on one of the computers on the real time lab.

## Running on the simulator

Start the [elevator simulator](https://github.com/TTK4145/Simulator-v2) with:
```
./simElevatorServer --port=<port>
```
Where `<port>` is an optional argument specifying the port to start the simulator on.

## Compile and run the code

First, make sure you are in the root directory and set the `GOPATH` variable (needs to be set every time you open a new terminal):
```
export GOPATH=$(pwd)
```

Install the code with the following command:
```
go install elevatorproject/main
```

Run the code with:
```
bin/main --id=<id> --port=<port> --addr=<addr>
```
Where all the command line arguments are optional.
`<id>` is the unique id of the elevator, i.e. an integer in the interval `[0, 255]`.
`<port>` and `<addr>` is the port and ip address of the elevator hardware you are connecting to.
If the elevator hardware/simulator is connected to the same computer as the elevator code is running on,
the default values should work.

### Example: running 3 simulators on the same computer
Start the simulators on different ports (in different terminals):
```
./simElevatorServer
./simElevatorServer --port=15677
./simElevatorServer --port=15777
```
Then run the elevator code (in different terminals):
```
bin/main
bin/main --id=1 --port=15677
bin/main --id=2 --port=15777
```
