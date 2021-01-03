// This file contains command-related structs and functions.
// Possible commands:
// 	1. send address message
// 	2. join introducer_address
// 	3. leave
// 	4. display member/id
// 	5. switch all-to-all/gossip
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// Command from user input
type Command struct {
	Method  string
	Payload []string
}

// read command from user and fill it into the channel
func readCommand(command chan Command) {
	// read input from user
	inputReader := bufio.NewReader(os.Stdin)
	fmt.Println("> Please enter a new command in the format of: 'Command Additional_info'")
	for {
		input, _ := inputReader.ReadString('\n')
		// fill input
		input = strings.TrimSpace(input)
		inputs := strings.Split(input, " ")
		command <- Command{inputs[0], inputs[1:]}
	}
}

// handle and dispatch command
func handleCommand(command Command) {
	switch command.Method {
	case "join":
		handleCommandJoin(command)
	case "leave":
		handleCommandLeave(command)
	case "send":
		handleCommandSend(command)
	case "switch":
		handleCommandSwitch(command)
	case "display":
		handleCommandDisplay(command)
	default:
		WarnLogger.Println("Unsupported Command!")
	}
}

// handle send command
// send command will send the given message to a given remote host
// ps: this is only for test purpose
func handleCommandSend(command Command) {
	addr, err := net.ResolveUDPAddr("udp", command.Payload[0])
	if err != nil {
		ErrorLogger.Println("Can't resolve address: ", err)
		os.Exit(1)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		ErrorLogger.Println("Can't dial: ", err)
		os.Exit(1)
	}
	defer conn.Close()

	DebugLogger.Println("Sent Command!")
	_, err = conn.Write([]byte(strings.Join(command.Payload[1:], " ")))
}

// handle join command
// this would send join command to a remote introducer host
func handleCommandJoin(command Command) {
	// introducer have no need to send join message
	if IntroducerMode { return }
	if len(command.Payload) == 0 {
		WarnLogger.Println("Invalid Join Arguments!")
		return
	}
	// send join to introducer node
	// if join vm
	if VMMode {
		vmNumber := command.Payload[0]
		vmPort := command.Payload[1]
		introducerAddrStr := fmt.Sprintf("fa20-cs425-g07-%s.cs.illinois.edu:%s", vmNumber, vmPort)
		sendMessage(Message{ Method: "JOIN" }, introducerAddrStr)
	} else { // if join other remote hosts
		introducerAddrStr := command.Payload[0]
		sendMessage(Message{ Method: "JOIN" }, introducerAddrStr)
	}
	InfoLogger.Println("Host", LocalUniqueID, "sent join request to the introducer.")
}

// handle leave command
func handleCommandLeave(command Command) {
	// send to all members
	broadcastMessage(Message{ Method: MSG_LEAVE })
	InfoLogger.Println("Host", LocalUniqueID, "left the system.")
	PrintBandwidthUsage()
	os.Exit(0)
}

// change between all-to-all and gossip
func handleCommandSwitch(command Command) {
	// set gossipMode to be !gossipMode
	GossipMode = !GossipMode

	broadcastMessage(Message{ Method: MSG_SWITCH })
	// empty all member's heartbeat
	//for _, member := range LocalMemberList {
	//	member.HeartbeatCounter = 0
	//}
	//handleMessage.sleep()
	InfoLogger.Println("Switched to another heartbeat style.")
}

// handle display command
// display member or id
func handleCommandDisplay(command Command) {
	if len(command.Payload) == 0 {
		WarnLogger.Println("Empty display argument!")
		return
	}
	switch command.Payload[0] {
	case "member":
		printMemberList()
	case "id":
		fmt.Println("The unique ID is:", LocalUniqueID)
	default:
		WarnLogger.Println("Invalid display argument!")
		break
	}
}
