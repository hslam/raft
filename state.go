package raft

type Role int

const (
	Follower	Role = 0
	Candidate	Role = 1
	Leader		Role = 2
)


type State interface{
	PreState()State
	NextState()State
	Update()
	State()Role
}

