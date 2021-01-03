// This file contains message-related structs and functions.
// Possible messages:
// 	1. ping id addr :
//		- all-to-all mode:
//			- reply with pong
// 			- no payload needed
// 		- gossip mode:
//			- no reply
//			- member list needs to be the payload
//  2. pong id addr : no reply, and update membership list
// 	3. join id addr : reply with pong
// 	3. leave id addr : no reply, and delete its entry in membership list
// 	5. switch all-to-all/gossip : no reply, switch its message type
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

// Message from other remote hosts
type Message struct {
	Method     string
	SenderID   string
	SenderAddr string
	Payload    []byte
}

const (
	MSG_PING   = "PING"
	MSG_PONG   = "PONG"
	MSG_JOIN   = "JOIN"
	MSG_LEAVE  = "LEAVE"
	MSG_SWITCH = "SWITCH"
)

// read from UDP continuously and file them into the channel
func readMessage(conn *net.UDPConn, messages chan Message) {
	if conn == nil {
		return
	}
	dataBuffer := make([]byte, MaxBufferSize)
	for {
		// receiver process
		cnt, _, err := conn.ReadFromUDP(dataBuffer)
		if err != nil {
			ErrorLogger.Println("failed to read from UDP:" + err.Error())
			return
		}
		// deserialize received message
		inMessageBytes := dataBuffer[0:cnt]
		inMessage := Message{}
		if err = json.Unmarshal(inMessageBytes, &inMessage); err != nil {
			ErrorLogger.Println("json unmarshal error:", err)
		}
		messages <- inMessage
		DebugLogger.Println("Message Received From", inMessage.SenderID)
	}
}

// send a message via UDP to a remote address
func sendMessage(outMessage Message, remoteAddrStr string) {
	// set the sender of message
	if outMessage.SenderID == "" { outMessage.SenderID = LocalUniqueID }
	if outMessage.SenderAddr == "" { outMessage.SenderAddr = LocalAddr }

	// send via UDP
	remoteAddr, err := net.ResolveUDPAddr("udp", remoteAddrStr)
	if err != nil {
		ErrorLogger.Println("Can't resolve address:", err)
		os.Exit(1)
	}
	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		ErrorLogger.Println("Can't dial:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// serialize message
	var messageBytes []byte
	if messageBytes, err = json.Marshal(outMessage); err != nil {
		ErrorLogger.Println("JSON marshal error:", err)
	}

	// simulate message loss
	if MessageLossRate > 0 {
		rand.Seed(time.Now().Unix())
		if rand.Float64() <= MessageLossRate {
			return  // like the message is lost
		}
	}

	// send message via udp
	_, err = conn.Write(messageBytes)
	if err != nil {
		ErrorLogger.Println("Failed to write to udp:", err.Error())
		os.Exit(1)
	}

	// added statistics
	BandwidthUsage += len(messageBytes)
	DebugLogger.Println("Message Sent To", remoteAddrStr)
}

// handle and dispatch received message
func handleMessage(message Message) {
	// check on input message----Message type
	switch message.Method {
	case MSG_PING: // ping, used for heartbeat
		handlePingMessage(message)
	case MSG_PONG: // pong, used for heartbeat
		handlePongMessage(message)
	case MSG_JOIN: // join, used for the join of a new machine
		handleJoinMessage(message)
	case MSG_LEAVE: // leave, used for the leave of a new machine
		handleLeaveMessage(message)
	case MSG_SWITCH:
		handleSwitchMessage(message) // switch to another heartbeat style
	default:
		WarnLogger.Println("Unsupported Message!")
	}
}

// handle ping message
func handlePingMessage(message Message) {
	if GossipMode { // if gossip mode
		if message.Payload == nil {  // gossip message should have payload
			InfoLogger.Println("A ping with a different heartbeating style is dropped. (Normal for switch)")
			return
		}
		var gossipMemberList []Member
		err := json.Unmarshal(message.Payload, &gossipMemberList)
		if err != nil {
			ErrorLogger.Println("JSON unmarshal error:", err)
		}
		// merge member list
		mergeGossipMemberList(gossipMemberList)
		DebugLogger.Println("Merged gossip from", message.SenderID, ".")
	} else { // if not gossip mode
		// when current process is all-to-all mode and others are gossip mode
		if message.Payload != nil {
			DebugLogger.Println("A ping with a different heartbeating style is dropped. (Normal for switch)")
			return
		}
		// update LocalMemberList
		heartbeatFromMember(message.SenderID, message.SenderAddr)
		//  reply with pong
		sendMessage(Message{Method: MSG_PONG}, message.SenderAddr)
	}

}

// handle pong message
func handlePongMessage(message Message) {
	// update LocalMemberList
	heartbeatFromMember(message.SenderID, message.SenderAddr)
}

// handle join message
// When a host received join, it would update its member list.
// If it is introducer, it would broadcast the message.
func handleJoinMessage(message Message) {
	// updated the new member in LocalMemberList
	updateMember(Member{
		ID:        message.SenderID,
		Addr:      message.SenderAddr,
		Status:    STAT_RUNNING,
		HeartbeatCounter:  1,
		Timestamp: time.Unix(time.Now().Unix(), 0),
	})

	// respond to the new member about self information
	sendMessage(Message{Method: MSG_PONG}, message.SenderAddr)

	// introducer would broadcast the message to all other active members in the group
	if IntroducerMode { broadcastMessage(message) }
}

// handle leave message
func handleLeaveMessage(message Message) {
	// receiver will change leaving process to "LEAVE"
	updatedMember := Member{
		ID:        message.SenderID,
		Addr:      message.SenderAddr,
		Status:    STAT_LEFT,
		Timestamp: time.Unix(time.Now().Unix(), 0),
	}
	updateMember(updatedMember)
	InfoLogger.Println("Process", message.SenderID, "left the system.")
}

// when a process receive a switch message, it would broadcast the switch to all its peer
func handleSwitchMessage(message Message) {
	// change Mode
	GossipMode = !GossipMode
	InfoLogger.Println("Process", LocalUniqueID, "has changed to another style.")
	// broadcast to all members
	//broadcastMessage(message)
	// empty all member's heartbeat
	for _, member := range LocalMemberList {
		member.HeartbeatCounter = 0
	}

}

// broadcast a message to all other members
func broadcastMessage(message Message, members ...Member) {
	if len(members) == 0 {
		members = LocalMemberList
	}
	for _, member := range members {
		// not send to oneself and left or failed host
		if member.ID == LocalUniqueID || member.Status == STAT_LEFT || member.Status == STAT_FAILED {
			continue
		}
		sendMessage(message, member.Addr)
	}
}

func messageToString(message Message) string {
	return fmt.Sprintf(
		"Method %s from host %d to %s, payload: %s",
		message.Method,
		message.SenderID,
		message.SenderAddr,
		string(message.Payload))
}