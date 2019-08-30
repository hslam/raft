package raft

import "time"

type Raft interface {
	Heartbeat(addr string) bool
	RequestVote(addr string) bool
	AppendEntries(addr string) bool
	InstallSnapshot(addr string) bool
	HandleRequestVote(req *RequestVoteRequest, res *RequestVoteResponse)error
	HandleAppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse)error
	HandleInstallSnapshot(req *InstallSnapshotRequest,res *InstallSnapshotResponse)error
}


type raft struct {
	node	*Node
	hearbeatTimeout			time.Duration
	requestVoteTimeout		time.Duration
	appendEntriesTimeout	time.Duration
	installSnapshotTimeout	time.Duration
}

func newRaft(node	*Node) Raft{
	return &raft{
		node:node,
		hearbeatTimeout:		DefaultHearbeatTimeout,
		requestVoteTimeout:		DefaultRequestVoteTimeout,
		appendEntriesTimeout:	DefaultAppendEntriesTimeout,
		installSnapshotTimeout:	DefaultInstallSnapshotTimeout,
	}
}
func (r *raft) Heartbeat(addr string) bool {
	var req =&AppendEntriesRequest{}
	req.Term=r.node.currentTerm
	req.LeaderId=r.node.leader
	req.LeaderCommit=r.node.commitIndex
	req.PrevLogIndex=r.node.commitIndex
	req.PrevLogTerm=r.node.currentTerm
	req.Entries=[]*Entry{}
	var ch =make(chan *AppendEntriesResponse)
	go func (rpcs *RPCs,addr string,req *AppendEntriesRequest, ch chan *AppendEntriesResponse) {
		var res =&AppendEntriesResponse{}
		if err := r.node.rpcs.CallAppendEntries(addr,req, res); err != nil {
			ch<-nil
		} else {
			ch <- res
		}
	}(r.node.rpcs,addr,req,ch)
	select {
	case res:=<-ch:
		if res==nil{
			Tracef("raft.Heartbeat %s -> %s error",r.node.address,addr)
			return false
		}
		//Tracef("Node.heartbeat %s -> %s",r.node.address,r.address)
		return true
	case <-time.After(r.hearbeatTimeout):
		Tracef("raft.Heartbeat %s -> %s time out",r.node.address,addr)
	}
	return false
}

func (r *raft) RequestVote(addr string) bool{
	var req =&RequestVoteRequest{}
	req.Term=r.node.currentTerm
	req.CandidateId=r.node.address
	req.LastLogIndex=r.node.commitIndex
	req.LastLogTerm=r.node.currentTerm
	var ch =make(chan *RequestVoteResponse)
	go func (rpcs *RPCs,addr string,req *RequestVoteRequest, ch chan *RequestVoteResponse) {
		var res =&RequestVoteResponse{}
		if err := rpcs.CallRequestVote(addr,req, res); err != nil {
			ch<-nil
		}else {
			ch <- res
		}
	}(r.node.rpcs,addr,req,ch)
	select {
	case res:=<-ch:
		if res==nil{
			Tracef("raft.RequestVote %s recv %s vote error",r.node.address,addr)
			return false
		}
		//Tracef("raft.RequestVote %s recv %s vote %t",r.node.address,addr,res.VoteGranted)
		if res.VoteGranted{
			r.node.votes.vote<-&Vote{candidateId:addr,vote:1,term:req.Term}
		}else {
			r.node.votes.vote<-&Vote{candidateId:addr,vote:0,term:req.Term}
		}
		return true

	case <-time.After(r.requestVoteTimeout):
		Tracef("raft.RequestVote %s recv %s vote time out",r.node.address,addr)
		r.node.votes.vote<-&Vote{candidateId:addr,vote:0,term:req.Term}
	}
	return false
}
func (r *raft) AppendEntries(addr string) bool{
	var req =&AppendEntriesRequest{}
	req.Term=r.node.currentTerm
	req.LeaderId=r.node.leader
	req.LeaderCommit=r.node.commitIndex
	req.PrevLogIndex=r.node.commitIndex
	req.PrevLogTerm=r.node.currentTerm
	req.Entries=[]*Entry{}
	var ch =make(chan *AppendEntriesResponse)
	go func (rpcs *RPCs,addr string,req *AppendEntriesRequest, ch chan *AppendEntriesResponse) {
		var res =&AppendEntriesResponse{}
		if err := r.node.rpcs.CallAppendEntries(addr,req, res); err != nil {
			ch<-nil
		} else {
			ch <- res
		}
	}(r.node.rpcs,addr,req,ch)
	select {
	case res:=<-ch:
		if res==nil{
			Tracef("raft.AppendEntries %s -> %s error",r.node.address,addr)
			return false
		}
		//Tracef("raft.AppendEntries %s -> %s",r.node.address,addr)
		return true
	case <-time.After(r.appendEntriesTimeout):
		Tracef("raft.AppendEntries %s -> %s time out",r.node.address,addr)
	}
	return false
}
func (r *raft) InstallSnapshot(addr string) bool{
	var req =&InstallSnapshotRequest{}
	req.Term=r.node.currentTerm
	var ch =make(chan *InstallSnapshotResponse)
	go func (rpcs *RPCs,addr string,req *InstallSnapshotRequest, ch chan *InstallSnapshotResponse) {
		var res =&InstallSnapshotResponse{}
		if err := r.node.rpcs.CallInstallSnapshot(addr,req, res); err != nil {
			ch<-nil
		}else {
			ch <- res
		}
	}(r.node.rpcs,addr,req,ch)
	select {
	case res:=<-ch:
		if res==nil{
			Tracef("raft.InstallSnapshot %s -> %s error",r.node.address,addr)
			return false
		}
		Tracef("raft.InstallSnapshot %s -> %s",r.node.address,addr)
		return true
	case <-time.After(r.installSnapshotTimeout):
		Tracef("raft.InstallSnapshot %s -> %s time out",r.node.address,addr)
	}
	return false
}

func (r *raft) HandleRequestVote(req *RequestVoteRequest, res *RequestVoteResponse)error {
	res.Term=r.node.currentTerm
	if req.Term<r.node.currentTerm{
		res.VoteGranted=false
		return nil
	}else if req.Term>r.node.currentTerm{
		r.node.currentTerm=req.Term
		r.node.stepDown()
	}
	if (r.node.votedFor==""||(r.node.votedFor==req.CandidateId&&r.node.currentTerm==req.Term))&&req.LastLogIndex>=r.node.lastLogIndex&&req.LastLogTerm>=r.node.lastLogTerm{
		res.VoteGranted=true
		r.node.votedFor=req.CandidateId
		r.node.stepDown()
	}
	return nil
}
func (r *raft) HandleAppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse)error {
	res.Term=r.node.currentTerm
	if req.Term<r.node.currentTerm{
		res.Success=false
		return nil
	}else if req.Term>r.node.currentTerm{
		r.node.currentTerm=req.Term
		r.node.stepDown()
	}
	if len(req.Entries)==0{
		if r.node.leader!=""&&r.node.leader!=req.LeaderId&&r.node.currentTerm!=req.Term{
			res.Success=false
			return nil
		}
		res.Success=true
		r.node.leader=req.LeaderId
		r.node.currentTerm=req.Term
		r.node.resetLastRPCTime()
		return nil
	}
	var lastEntryIndex uint64
	for _,v :=range req.Entries {
		lastEntryIndex=v.Index
	}
	if req.LeaderCommit>r.node.commitIndex{
		r.node.commitIndex=minUint64(req.LeaderCommit,lastEntryIndex)
		res.Success=true
		return nil
	}
	return nil
}
func (r *raft) HandleInstallSnapshot(req *InstallSnapshotRequest,res *InstallSnapshotResponse)error {
	return nil
}



