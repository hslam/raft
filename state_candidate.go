package raft

type CandidateState struct{
}
func (this *CandidateState)PreState()State{
	return new(FollowerState)
}
func (this *CandidateState)NextState()State{
	return new(LeaderState)
}

func (this *CandidateState) Update(){
}
func (this *CandidateState) State()Role{
	return Candidate
}