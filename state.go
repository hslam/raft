package raft


const (
	Follower	= "Follower"
	Candidate	= "Candidate"
	Leader		= "Leader"
)


type State interface{
	Init()
	Update()
	String()string
	StepDown()State
	NextState()State
}

