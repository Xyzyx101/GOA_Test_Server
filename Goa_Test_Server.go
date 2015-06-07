package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	//"bytes"
)

type Command struct {
	PlayerId     int
	CommandType  int
	Time         float64
	PositionX    float64
	PositionY    float64
	PositionZ    float64
	MoveForward  float64
	MoveSideways float64
	Jump         bool
	Crouch       bool
}

type CommandType int

const (
	MOVE         = 0
	CONNECT      = 1
	DISCONNECT   = 2
	SPAWN_LOCAL  = 3
	SPAWN_REMOTE = 4
)

type Players map[int]*net.UDPAddr

var players = make(map[int]*net.UDPAddr, 4)

func main() {
	port := "8042"
	ServerAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%s", port))
	if err != nil {
		fmt.Println("Resolve Address Error: ", err)
		os.Exit(0)
	} else {
		fmt.Printf("Server connected at port: %s\n", port)
	}

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	defer ServerConn.Close()
	if err != nil {
		fmt.Println("ListenUDP Error: ", err)
	}

	buf := make([]byte, 1024)

	nextPlayerId := playerIdIncrementer()

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr, "  ", n, " bytes")
		command, err := extractCommandFromJSON(buf[0:n])
		if err != nil {
			fmt.Println("JSON Error: ", err)
		} else {
			switch command.CommandType {
			case MOVE:
				for playerId, playerAddr := range players {
					n, err := sendCommand(*command, ServerConn, playerAddr)
					if err != nil {
						fmt.Println("Error sending move: ", err)
					}
					fmt.Println(n, " bytes sent to playerId:", playerId)
				}
			case CONNECT:
				spawnPlayerId := nextPlayerId()
				command.PlayerId = spawnPlayerId
				players[command.PlayerId] = addr

				// Spawn connecting player at all clients
				for playerId, playerAddr := range players {
					if playerId == spawnPlayerId {
						command.CommandType = SPAWN_LOCAL
					} else {
						command.CommandType = SPAWN_REMOTE
					}
					n, err := sendCommand(*command, ServerConn, playerAddr)
					if err != nil {
						fmt.Println("Error sending local spawn: ", err)
					} else {
						fmt.Println(n, " bytes sent to playerId:", spawnPlayerId)
					}
				}

				// spawn other players at new connection
				command.CommandType = SPAWN_REMOTE
				for playerId, _ := range players {
					if playerId == spawnPlayerId {
						continue
					}
					command.PlayerId = playerId
					n, err := sendCommand(*command, ServerConn, addr)
					if err != nil {
						fmt.Println("Error sending remote spawn: ", err)
					} else {
						fmt.Println(n, " bytes sent to playerId:", playerId)
					}
				}
			}
		}
		if err != nil {
			fmt.Println("JSON Error: ", err)
		}
	}
}

func sendCommand(command Command, conn *net.UDPConn, addr *net.UDPAddr) (n int, err error) {
	buf, err := json.Marshal(command)
	if err != nil {
		return 0, err
	}
	n, err = conn.WriteToUDP(buf, addr)
	if err != nil {
		return 0, err
	}
	return
}

func extractCommandFromJSON(buf []byte) (command *Command, err error) {
	//var command Command
	err = json.Unmarshal(buf, &command)
	if err != nil {
		return nil, err
	}
	fmt.Println("PlayerId:", command.PlayerId, " CommandType:", command.CommandType)
	return
}

func playerIdIncrementer() func() int {
	playerId := 0
	return func() int {
		playerId += 1
		return playerId
	}
}
