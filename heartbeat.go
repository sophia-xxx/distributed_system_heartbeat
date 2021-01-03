// This file includes functions for heartbeating style of failure detection.
// Two variants:
//	1. All to All heartbeating
//	2. Gossip heartbeating: Push-based Gossip
package main

import (
	"encoding/json"
	"log"
	"time"
)

// check whether any process failed
func CheckFailure() {
	// check timeout(failure) of all members
	for _, member := range LocalMemberList {
		if member.ID == LocalUniqueID { continue }
		// otherwise, check time span between now and the last time receive heartbeat
		timeSpan := time.Now().Sub(member.Timestamp)
		// check failed process and clean up
		if member.Status == STAT_FAILED || member.Status == STAT_LEFT {
			if timeSpan > CleanUpSeconds * time.Second { removeMember(member) }
		} else {
			var TimeOutSeconds time.Duration
			if GossipMode {
				TimeOutSeconds = GossipTimeOutSeconds
			} else {
				TimeOutSeconds = AllToAllTimeOutSeconds
			}
			if timeSpan > TimeOutSeconds * time.Second {
				updateMember(Member{
					ID:        member.ID,
					Addr:      member.Addr,
					Status:    STAT_FAILED,
					Timestamp: time.Unix(time.Now().Unix(), 0),
				})
				InfoLogger.Println("Host", member.ID, "failed.")
			}
		}
	}
}

// periodically send out heartbeat
func RunHeartBeat() {
	// set ticker to heartbeat periodically
	ticker := time.NewTicker(HeartbeatPeriod * time.Millisecond)
	for range ticker.C {
		if GossipMode {
			gossipHeartBeat()
		} else {
			allToAllHeartBeat()
		}
	}
}

// broadcast heartbeat to all peers
func allToAllHeartBeat() {
	// send PING message to all RUNNING process
	broadcastMessage(Message{Method: MSG_PING})
}

// GossipMode style heartbeat, send LocalMemberList
func gossipHeartBeat() {
	getMemberById(LocalUniqueID).HeartbeatCounter++
	// serialize LocalMemberList
	memberListBytes, err := json.Marshal(LocalMemberList)
	if err != nil {
		log.Fatal("json marshal error:", err)
	}
	// send GOSSIP message to completely random processes
	broadcastMessage(Message{Method: MSG_PING, Payload: memberListBytes}, getRandomMembers(GossipRate)...)
}