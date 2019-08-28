package raft


type FollowerState struct{
	c *Context
}
func (this *FollowerState)PreState()State{
	return this
}
func (this *FollowerState)NextState()State{
	return new(CandidateState)
}

func (this *FollowerState) Update(){
}

func (this *FollowerState) State()Role{
	return Follower
}