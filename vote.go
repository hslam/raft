package raft

import (
	"strconv"
	"time"
	"sync"
)

type Vote struct {
	id	string
	term		uint64
	vote		int
}
func newVote(id	string,term	uint64,vote	int) *Vote {
	v:=&Vote{
		id:id,
		term:term,
		vote:vote,
	}
	return v
}
func (v *Vote)Key() string {
	return v.id+strconv.FormatUint(v.term,10)
}

type Votes struct {
	mu								sync.RWMutex
	node 							*Node
	vote	 						chan *Vote
	voteDic	 						map[string]int
	voteTotal	 					int
	voteCount	 					int
	quorum							int
	total							int
	notice							chan bool
	timeout							time.Duration
}

func newVotes(node *Node) *Votes {
	votes:=&Votes{
		node:node,
		notice:make(chan bool,1),
	}
	votes.Reset(1)
	votes.Clear()
	return votes
}

func (votes *Votes)AddVote(v *Vote) {
	votes.mu.Lock()
	defer votes.mu.Unlock()
	if _,ok:=votes.voteDic[v.Key()];ok{
		return
	}
	votes.voteDic[v.Key()]=v.vote
	if v.term==votes.node.currentTerm.Id(){
		votes.voteTotal+=1
		votes.voteCount+=v.vote
	}
	if votes.quorum>0&&votes.voteCount>=votes.quorum{
		votes.notice<-true
	}else if votes.total>0&&votes.voteTotal>=votes.total{
		votes.notice<-false
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
func (votes *Votes)SetQuorum(quorum int){
	votes.quorum=quorum
}
func (votes *Votes)SetTotal(total int){
	votes.total=total
}
func (votes *Votes)GetNotice()chan bool{
	 return votes.notice
}
func (votes *Votes)SetTimeout(timeout time.Duration){
	votes.timeout=timeout
	go votes.run()
}
func (votes *Votes)run() {
	select {
	case <-time.After(votes.timeout):
		votes.notice<-false
	}
}