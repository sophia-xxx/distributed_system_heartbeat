// This file is used to initialize and run the main program.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"
)

// entry point
func main() {
	initialize()
	// run the daemon
	DaemonRun()
	// Exit
	os.Exit(0)
}

// initialize
func initialize()  {
	// parse flags
	var localhost, localport string
	flag.StringVar(&localhost, "host", "localhost", "the local host")
	flag.StringVar(&localport, "port", "2333", "the local port")
	flag.BoolVar(&VMMode, "vm", false, "whether run in the vm")
	flag.BoolVar(&IntroducerMode, "introducer", false, "whether is the introducer server")
	flag.BoolVar(&DebugMode, "debug", false, "whether is in debug mode")
	flag.BoolVar(&GossipMode, "gossip", false, "whether is in gossip mode")
	flag.Float64Var(&MessageLossRate, "experiment", 0, "whether simulate message loss")
	flag.Parse()

	// if not in debug mode, discard debug output
	if !DebugMode {
		DebugLogger.SetOutput(ioutil.Discard)
	}
	BandwidthUsage = 0
	// initialize local address
	if VMMode {
		LocalAddr = fmt.Sprintf("fa20-cs425-g07-%s.cs.illinois.edu:%s", localhost, localport)
	} else {
		LocalAddr = localhost + ":" + localport
	}

	// initialize LocalUniqueID and membership list
	LocalUniqueID = generateUniqueId()
	initializeMemberInfo(LocalUniqueID, LocalAddr)

	// check initialization
	DebugLogger.Println("Check initialization:", LocalAddr, IntroducerMode, LocalUniqueID, LocalMemberList)
}

// run background server daemon
//	host, post: the address where server would be listening on
// 	userCommand: a channel for user input, such as join, leave, switch
func DaemonRun() {
	// resolve the udp server address
	serverAddr, err := net.ResolveUDPAddr("udp", LocalAddr)
	if err != nil {
		ErrorLogger.Println("Can't resolve address:", err.Error())
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		ErrorLogger.Println("Failed to start server:", err.Error())
	}
	// output server info
	InfoLogger.Println("Server Started at:", serverAddr.String(),
		", Unique ID:", LocalUniqueID,
		", IntroducerMode:", IntroducerMode,
		", DebugMode:", DebugMode,
		", GossipMode:", GossipMode)
	// close connect after finishing main
	defer conn.Close()

	// create channels for taking messages and commands
	messages := make(chan Message, 10)
	go readMessage(conn, messages) // read messages from UDP
	command := make(chan Command)
	go readCommand(command) // read command from user

	go RunHeartBeat()

	// continuously, handle messages and commands, and
	for {
		CheckFailure()
		select {
		case message := <-messages:
			handleMessage(message)
		case command := <-command:
			handleCommand(command)
		default:
			time.Sleep(SleepPeriod * time.Millisecond)
		}
	}
}

// print the Bandwidth Usage
func PrintBandwidthUsage() {
	InfoLogger.Println("Bandwidth Used:", BandwidthUsage, "Bytes.")
}