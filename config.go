package raft

import (
	"time"
)


const (
	DefaultHearbeatTicker = 50 * time.Millisecond
	DefaultElectionTimeout = 150 * time.Millisecond

)

type Configuration struct {
	//CommitIndex uint64 	`json:"CommitIndex"`
	Nodes []string		`json:"Nodes"`
}




