package raft

import "time"

type Raft interface {
	Heartbeat(addr string,prevLogIndex,prevLogTerm uint64) (bool,bool)
	RequestVote(addr string) bool
	AppendEntries(addr string,prevLogIndex,prevLogTerm uint64,entries []*Entry)  (bool,bool)
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


func (r *raft) Heartbeat(addr string,prevLogIndex,prevLogTerm uint64) (bool,bool) {
	var req =&AppendEntriesRequest{}
	req.Term=r.node.currentTerm.Id()
	req.LeaderId=r.node.leader
	req.LeaderCommit=r.node.commitIndex
	req.PrevLogIndex=prevLogIndex
	req.PrevLogTerm=prevLogTerm
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
			return false,false
		}
		if res.Term>r.node.currentTerm.Id(){
			r.node.currentTerm.Set(res.Term)
			r.node.stepDown()
		}
		//Tracef("Node.heartbeat %s -> %s",r.node.address,r.address)
		if res.Success{
			return true,true
		}
		return true,false
	case <-time.After(r.hearbeatTimeout):
		Tracef("raft.Heartbeat %s -> %s time out",r.node.address,addr)
	}
	return false,false
}

func (r *raft) RequestVote(addr string) (bool){
	var req =&RequestVoteRequest{}
	req.Term=r.node.currentTerm.Id()
	req.CandidateId=r.node.address
	req.LastLogIndex=r.node.lastLogIndex
	req.LastLogTerm=r.node.lastLogTerm
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
		if res.Term>r.node.currentTerm.Id(){
			r.node.currentTerm.Set(res.Term)
			r.node.stepDown()
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
func (r *raft) AppendEntries(addr string,prevLogIndex,prevLogTerm uint64,entries []*Entry)  (bool,bool){
	var req =&AppendEntriesRequest{}
	req.Term=r.node.currentTerm.Id()
	req.LeaderId=r.node.leader
	req.LeaderCommit=r.node.commitIndex
	req.PrevLogIndex=prevLogIndex
	req.PrevLogTerm=prevLogTerm
	req.Entries=entries
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
			return false,false
		}
		if res.Term>r.node.currentTerm.Id(){
			r.node.currentTerm.Set(res.Term)
			r.node.stepDown()
			return true,false
		}
		if res.Success{
			return true,true
		}
		//Tracef("raft.AppendEntries %s -> %s",r.node.address,addr)
		return true,false
	case <-time.After(r.appendEntriesTimeout):
		Tracef("raft.AppendEntries %s -> %s time out",r.node.address,addr)
	}
	return false,false
}
func (r *raft) InstallSnapshot(addr string) bool{
	var req =&InstallSnapshotRequest{}
	req.Term=r.node.currentTerm.Id()
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
		if res.Term>r.node.currentTerm.Id(){
			r.node.currentTerm.Set(res.Term)
			r.node.stepDown()
		}
		Tracef("raft.InstallSnapshot %s -> %s",r.node.address,addr)
		return true
	case <-time.After(r.installSnapshotTimeout):
		Tracef("raft.InstallSnapshot %s -> %s time out",r.node.address,addr)
	}
	return false
}

func (r *raft) HandleRequestVote(req *RequestVoteRequest, res *RequestVoteResponse)error {
	res.Term=r.node.currentTerm.Id()
	if req.Term<r.node.currentTerm.Id(){
		res.VoteGranted=false
		return nil
	}else if req.Term>r.node.currentTerm.Id(){
		r.node.currentTerm.Set(req.Term)
		r.node.stepDown()
		if r.node.state.String()==Leader||r.node.state.String()==Candidate{
			return nil
		}
	}
	if (r.node.votedFor==""||r.node.votedFor==req.CandidateId)&&req.LastLogIndex>=r.node.lastLogIndex&&req.LastLogTerm>=r.node.lastLogTerm{
		res.VoteGranted=true
		r.node.votedFor=req.CandidateId
		r.node.stepDown()
	}
	return nil

}
func (r *raft) HandleAppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse)error {
	res.Term=r.node.currentTerm.Id()
	if req.Term<r.node.currentTerm.Id(){
		res.Success=false
		return nil
	}else if req.Term>r.node.currentTerm.Id(){
		r.node.currentTerm.Set(req.Term)
		r.node.stepDown()
		if r.node.state.String()==Leader||r.node.state.String()==Candidate{
			return nil
		}
	}
	if len(req.Entries)==0{
		if 	r.node.leader==""{
			r.node.leader=req.LeaderId
		}
		if r.node.leader==req.LeaderId{
			res.Success=true
			r.node.election.Reset()
		}else {
			res.Success=false
			if r.node.state.String()==Leader||r.node.state.String()==Candidate{
				Tracef("raft.HandleAppendEntries %s State %s Term %d",r.node.address, r.node.State(),r.node.currentTerm.Id())
			}
			r.node.stepDown()
		}
		return nil
	}
	if req.PrevLogIndex!=r.node.prevLogIndex||req.PrevLogTerm!=r.node.prevLogTerm{
		res.Success=false
		return nil
	}
	r.node.log.appendEntries(req.Entries)
	var lastEntryIndex =req.Entries[len(req.Entries)-1].Index
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



