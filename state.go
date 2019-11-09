package raft

const (
	Follower	= "Follower"
	Candidate	= "Candidate"
	Leader		= "Leader"
)

type State interface{
	Start()
	Update()bool
	FixedUpdate()
	String()string
	StepDown()State
	NextState()State
}

