# Elevator project

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
