package raft

const (
	Follower	= "Follower"
	Candidate	= "Candidate"
	Leader		= "Leader"
)

type State interface{
	Reset()
	Update()
	FixedUpdate()
	String()string
	StepDown()State
	NextState()State
}

