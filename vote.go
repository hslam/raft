package raft

import "strconv"

type Vote struct {
	candidateId	string
	term		uint64
	vote		int
}

func (v *Vote)Key() string {
	return v.candidateId+strconv.FormatUint(v.term,10)
}

type Votes struct {
	node 							*Node
	vote	 						chan *Vote
	voteDic	 						map[string]int
	voteTotal	 					int
	voteCount	 					int
	quorum							int
}

func newVotes(node *Node) *Votes {
	votes:=&Votes{
		node:node,
	}
	votes.Reset(1)
	votes.Clear()
	return votes
}

func (votes *Votes)AddVote(v *Vote) {
	if _,ok:=votes.voteDic[v.Key()];ok{
		return
	}
	votes.voteDic[v.Key()]=v.vote
	if v.term==votes.node.currentTerm.Id(){
		votes.voteTotal+=1
		votes.voteCount+=v.vote
	}
}

func (votes *Votes)Clear() {
	for{
		if len(votes.vote)>0{
			<-votes.vote
		}else {
			break
		}
	}
	votes.voteDic=make(map[string]int)
	votes.voteCount=0
	votes.voteTotal=0
}
func (votes *Votes)Reset(count int) {
	if votes.vote!=nil{
		close(votes.vote)
	}
	votes.vote=make(chan *Vote,count)
}

func (votes *Votes)Count()int {
	return votes.voteCount
}
func (votes *Votes)Total()int {
	return votes.voteTotal
}