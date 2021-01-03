// This file contains global variables.
package main

import (
	"log"
	"os"
)

// some const parameters
const (
	MaxBufferSize = 4096 // max size of buffers
	SleepPeriod   = 50   // period of sleep when there is no task to do
	// failure related:
	GossipTimeOutSeconds  = 10   // max timeouts in seconds
	AllToAllTimeOutSeconds  = 5  // max timeouts in seconds
	CleanUpSeconds  = 600   // time for cleaning up failed processes in seconds
	HeartbeatPeriod = 1000 // period of sending out ping in nanoseconds
	// gossip related
	GossipRate = 5 // how many times a gossip would be transferred to
)

var VMMode bool         // whether run in vm
var DebugMode bool      // whether debug mode
var GossipMode bool     // whether in gossip mode
var IntroducerMode bool // whether is the introducer

// membership list
var LocalMemberList []Member // list storing all info about members

// define new process LocalUniqueID
var LocalUniqueID string

// define local host and port
var LocalAddr string

// loggers
var (
	InfoLogger  = log.New(os.Stdout, "[info ]", log.Ltime)
	DebugLogger = log.New(os.Stderr, "[debug]", log.Ltime)
	WarnLogger  = log.New(os.Stderr, "[warn ]", log.Ltime)
	ErrorLogger	= log.New(os.Stderr, "[error]", log.Ltime)
)

// for statistics and experiment
var (
	BandwidthUsage  int     // in bytes
	MessageLossRate float64 // rate of message loss
)