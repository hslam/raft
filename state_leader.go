package raft

type LeaderState struct{
	c *Context
	NextIndex		[]uint64
	MatchIndex		[]uint64
}
func (this *LeaderState)PreState()State{
	return new(FollowerState)
}
func (this *LeaderState)NextState()State{
	return new(FollowerState)
}

func (this *LeaderState) Update(){
}
func (this *LeaderState) State()Role{
	return Leader
}

