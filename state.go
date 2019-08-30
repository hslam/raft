package raft


const (
	Follower	= "Follower"
	Candidate	= "Candidate"
	Leader		= "Leader"
)


type State interface{
	PreState()State
	NextState()State
	Init()
	Update()
	String()string
}

