// This file contains member-related structs and functions.
package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Member struct {
	ID               string
	Addr             string // IP address
	HeartbeatCounter int    // heartbeat sequence
	Status           string // running, left, failed
	Timestamp        time.Time
}

const (
	STAT_RUNNING = "Running" // running
	STAT_LEFT    = "Left"    // left
	STAT_FAILED  = "Failed"  // failed
)

// generate an unique ID
func generateUniqueId() string {
	// id that includes a timestamp and IP address
	return fmt.Sprintf(
		"%d@%s",
		time.Now().Unix()%(1000000), // last six digits of timestamp
		LocalAddr)
}

// print a single membership
func printMember(m Member) {
	fmt.Printf("  - %s, status: %s, timestamp: %s, address: %s, heartbeat counter: %d\n", m.ID, m.Status, m.Timestamp.Format("2006-01-02 15:04:05"), m.Addr, m.HeartbeatCounter)
}

// print the membership list
func printMemberList() {
	fmt.Printf("MemberList: \n")
	// sort the member list, not work
	// sort.Slice(LocalMemberList, func(i, j int) bool {
	// 	return LocalMemberList[i].ID < LocalMemberList[j].ID
	// })
	for _, m := range LocalMemberList {
		//if m.Status == STAT_RUNNING
			printMember(m)
	}
	if GossipMode {
		InfoLogger.Println("Current Membership Mode: Gossip Style.")
	} else {
		InfoLogger.Println("Current Membership Mode: All-to-All Style.")
	}
}

// initialize the local member list
func initializeMemberInfo(id string, addr string) {
	membershipList := make([]Member, 1)
	timestamp := time.Unix(time.Now().Unix(), 0)
	membershipList[0] = Member{
		ID:               id,
		Status:           STAT_RUNNING,
		HeartbeatCounter: 0,
		Timestamp:        timestamp,
		Addr:             addr,
	}
	LocalMemberList = membershipList
	DebugLogger.Printf("Init memberlist success.\n")
}

func getMemberById(id string) *Member {
	for ind := range LocalMemberList {
		if LocalMemberList[ind].ID == id {
			return &LocalMemberList[ind]
		}
	}
	return nil
}

// whether the member is an active remote host
func isValidRemoteMember(member Member) bool {
	return member.ID != LocalAddr && member.Status == STAT_RUNNING
}

// return requiredSize active members
func getRandomMembers(requiredSize int) []Member {
	if len(LocalMemberList) <= requiredSize-1 {
		return LocalMemberList
	} else {
		tmpList := make([]Member, len(LocalMemberList))
		copy(tmpList, LocalMemberList)
		// shuffle the slice
		rand.Seed(time.Now().Unix())
		rand.Shuffle(len(tmpList), func(i, j int) { tmpList[i], tmpList[j] = tmpList[j], tmpList[i] })

		resultList := make([]Member, requiredSize)
		for _, member := range tmpList {
			if isValidRemoteMember(member) {
				resultList = append(resultList, member)
			}
		}
		return resultList
	}
}

// merge membership list
func mergeGossipMemberList(newMemberList []Member) {
	for _, member := range newMemberList {
		// if is itself
		if member.ID == LocalUniqueID { continue }
		// search its corresponding member in the list
		oldMember := getMemberById(member.ID)
		// if not found
		if oldMember == nil {
			// only insert a new member if it is active
			if isValidRemoteMember(member) {
				insertMember(member.ID, member.Addr)
			}
			continue
		}
		// compare, if outdated, update the entry
		if oldMember.HeartbeatCounter < member.HeartbeatCounter {
			DebugLogger.Println("Updated the member:", member.ID)
			oldMember.HeartbeatCounter = member.HeartbeatCounter
			oldMember.Timestamp = time.Unix(time.Now().Unix(), 0)
		}
	}
}

// renew a member when receiving a ping or pong
func heartbeatFromMember(heartbeatID string, heartbeatAddrStr string) {
	found := false
	for i := range LocalMemberList {
		if LocalMemberList[i].ID == heartbeatID {
			LocalMemberList[i].Addr = heartbeatAddrStr
			LocalMemberList[i].HeartbeatCounter++
			LocalMemberList[i].Timestamp = time.Unix(time.Now().Unix(), 0)
			found = true
			break
		}
	}
	if !found {
		insertMember(heartbeatID, heartbeatAddrStr)
	}
}

// insert a new member into the member list
func insertMember(newMemberID string, newMemberAddrStr string) {
	LocalMemberList = append(LocalMemberList, Member{
		ID:               newMemberID,
		Addr:             newMemberAddrStr,
		Status:           STAT_RUNNING,
		HeartbeatCounter: 1,
		Timestamp:        time.Unix(time.Now().Unix(), 0),
	})
	InfoLogger.Println("Member", newMemberID, "is added into the member list.")
}

// update a member in the member list. If the member is not in the member list, insert it.
func updateMember(newMember Member) {
	find := false
	for index := range LocalMemberList {
		if LocalMemberList[index].ID == newMember.ID {
			find = true
			LocalMemberList[index] = newMember
			break
		}
	}
	if !find {
		insertMember(newMember.ID, newMember.Addr)
	}
}

// remove a member from the member list
func removeMember(oldMember Member) {
	var memberIndex int
	for index := range LocalMemberList {
		if LocalMemberList[index].ID == oldMember.ID {
			memberIndex = index
			break
		}
	}
	LocalMemberList = append(LocalMemberList[:memberIndex], LocalMemberList[memberIndex+1:]...)
	InfoLogger.Println("Member", oldMember.ID, "is removed from the member list.")
}
