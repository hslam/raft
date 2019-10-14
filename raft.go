package raft

import "time"

type Raft interface {
	Hearbeat(addr string,prevLogIndex,prevLogTerm uint64)  (nextIndex uint64,term uint64,success bool,ok bool)
	RequestVote(addr string) (ok bool)
	AppendEntries(addr string,prevLogIndex,prevLogTerm uint64,entries []*Entry)  (nextIndex uint64,term uint64,success bool,ok bool)
	InstallSnapshot(addr string) (ok bool)
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

func (r *raft) Hearbeat(addr string,prevLogIndex,prevLogTerm uint64)  (nextIndex uint64,term uint64,success bool,ok bool){
	var req =&AppendEntriesRequest{}
	req.Term=r.node.currentTerm.Id()
	req.LeaderId=r.node.leader
	req.LeaderCommit=r.node.commitIndex.Id()
	req.PrevLogIndex=prevLogIndex
	req.PrevLogTerm=prevLogTerm
	req.Entries=[]*Entry{}
	var ch =make(chan *AppendEntriesResponse,1)
	var errCh =make(chan error,1)
	go func (rpcs *RPCs,addr string,req *AppendEntriesRequest, ch chan *AppendEntriesResponse,errCh chan error) {
		var res =&AppendEntriesResponse{}
		if err := r.node.rpcs.CallAppendEntries(addr,req, res); err != nil {
			errCh<-err
		} else {
			ch <- res
		}
	}(r.node.rpcs,addr,req,ch,errCh)
	select {
	case res:=<-ch:
		if res.Term>r.node.currentTerm.Id(){
			r.node.currentTerm.Set(res.Term)
			r.node.stepDown()
		}
		//Tracef("raft.Hearbeat %s -> %s",r.node.address,addr)
		return res.NextIndex,res.Term,res.Success,true
	case err:=<-errCh:
		Tracef("raft.Hearbeat %s -> %s error %s",r.node.address,addr,err.Error())
		return 0,0,false,false
	case <-time.After(r.hearbeatTimeout):
		Tracef("raft.Hearbeat %s -> %s time out",r.node.address,addr)
	}
	return 0,0,false,false
}
func (r *raft) RequestVote(addr string) (ok bool){
	var req =&RequestVoteRequest{}
	req.Term=r.node.currentTerm.Id()
	req.CandidateId=r.node.address
	req.LastLogIndex=r.node.lastLogIndex
	req.LastLogTerm=r.node.lastLogTerm
	var ch =make(chan *RequestVoteResponse,1)
	var errCh =make(chan error,1)
	go func (rpcs *RPCs,addr string,req *RequestVoteRequest, ch chan *RequestVoteResponse,errCh chan error) {
		var res =&RequestVoteResponse{}
		if err := rpcs.CallRequestVote(addr,req, res); err != nil {
			errCh<-err
		}else {
			ch <- res
		}
	}(r.node.rpcs,addr,req,ch,errCh)
	select {
	case res:=<-ch:
		if res.Term>r.node.currentTerm.Id(){
			r.node.currentTerm.Set(res.Term)
			r.node.stepDown()
		}
		//Tracef("raft.RequestVote %s recv %s vote %t",r.node.address,addr,res.VoteGranted)
		if res.VoteGranted{
			r.node.votes.vote<-newVote(addr,req.Term,1)
		}else {
			r.node.votes.vote<-newVote(addr,req.Term,0)
		}
		return true
	case err:=<-errCh:
		Tracef("raft.RequestVote %s recv %s vote error %s",r.node.address,addr,err.Error())
		return false
	case <-time.After(r.requestVoteTimeout):
		Tracef("raft.RequestVote %s recv %s vote time out",r.node.address,addr)
		r.node.votes.vote<-newVote(addr,req.Term,0)
	}
	return false
}
func (r *raft) AppendEntries(addr string,prevLogIndex,prevLogTerm uint64,entries []*Entry)  (nextIndex uint64,term uint64,success bool,ok bool){
	var req =&AppendEntriesRequest{}
	req.Term=r.node.currentTerm.Id()
	req.LeaderId=r.node.leader
	req.LeaderCommit=r.node.commitIndex.Id()
	req.PrevLogIndex=prevLogIndex
	req.PrevLogTerm=prevLogTerm
	req.Entries=entries
	var ch =make(chan *AppendEntriesResponse,1)
	var errCh =make(chan error,1)
	go func (rpcs *RPCs,addr string,req *AppendEntriesRequest, ch chan *AppendEntriesResponse,errCh chan error) {
		var res =&AppendEntriesResponse{}
		if err := r.node.rpcs.CallAppendEntries(addr,req, res); err != nil {
			errCh<-err
		} else {
			ch <- res
		}
	}(r.node.rpcs,addr,req,ch,errCh)
	select {
	case res:=<-ch:
		if res.Term>r.node.currentTerm.Id(){
			r.node.currentTerm.Set(res.Term)
			r.node.stepDown()
			if len(entries)>0{
				return res.NextIndex,res.Term,false,true
			}
		}
		//Tracef("raft.AppendEntries %s -> %s",r.node.address,addr)
		return res.NextIndex,res.Term,res.Success,true
	case err:=<-errCh:
		Tracef("raft.AppendEntries %s -> %s error %s",r.node.address,addr,err.Error())
		return 0,0,false,false
	case <-time.After(r.appendEntriesTimeout):
		Tracef("raft.AppendEntries %s -> %s time out",r.node.address,addr)
	}
	return 0,0,false,false
}
func (r *raft) InstallSnapshot(addr string) bool{
	var req =&InstallSnapshotRequest{}
	req.Term=r.node.currentTerm.Id()
	var ch =make(chan *InstallSnapshotResponse,1)
	var errCh =make(chan error,1)
	go func (rpcs *RPCs,addr string,req *InstallSnapshotRequest, ch chan *InstallSnapshotResponse,errCh chan error) {
		var res =&InstallSnapshotResponse{}
		if err := r.node.rpcs.CallInstallSnapshot(addr,req, res); err != nil {
			errCh<-err
		}else {
			ch <- res
		}
	}(r.node.rpcs,addr,req,ch,errCh)
	select {
	case res:=<-ch:
		if res.Term>r.node.currentTerm.Id(){
			r.node.currentTerm.Set(res.Term)
			r.node.stepDown()
		}
		Tracef("raft.InstallSnapshot %s -> %s",r.node.address,addr)
		return true
	case err:=<-errCh:
		Tracef("raft.InstallSnapshot %s -> %s error %s",r.node.address,addr,err.Error())
		return false
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
	}else if req.Term>r.node.currentTerm.Id()&&req.LastLogIndex>=r.node.lastLogIndex&&req.LastLogTerm>=r.node.lastLogTerm{
		r.node.currentTerm.Set(req.Term)
		r.node.votedFor.Reset()
		r.node.stepDown()
	}
	if r.node.state.String()==Leader||r.node.state.String()==Candidate{
		res.VoteGranted=false
		return nil
	} else if (r.node.votedFor.String()==""||r.node.votedFor.String()==req.CandidateId)&&req.LastLogIndex>=r.node.lastLogIndex&&req.LastLogTerm>=r.node.lastLogTerm{
		res.VoteGranted=true
		r.node.votedFor.Set(req.CandidateId)
		r.node.stepDown()
	}
	return nil

}
func (r *raft) HandleAppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse)error {
	res.Term=r.node.currentTerm.Id()
	res.NextIndex=r.node.nextIndex
	if req.Term<r.node.currentTerm.Id(){
		res.Success=false
		return nil
	}else if req.Term>r.node.currentTerm.Id(){
		r.node.currentTerm.Set(req.Term)
		r.node.stepDown()
		if r.node.state.String()==Leader||r.node.state.String()==Candidate{
			res.Success=false
			return nil
		}
	}
	if 	r.node.leader==""{
		r.node.votedFor.Set(req.LeaderId)
		r.node.leader=req.LeaderId
		Tracef("raft.HandleAppendEntries %s State:%s leader-%s Term:%d",r.node.address, r.node.State(),r.node.leader,r.node.currentTerm.Id())
	}
	if r.node.leader!=req.LeaderId{
		res.Success=false
		r.node.stepDown()
		return nil
	}
	if r.node.state.String()==Leader||r.node.state.String()==Candidate{
		r.node.stepDown()
		Tracef("raft.HandleAppendEntries %s State:%s Term:%d",r.node.address, r.node.State(),r.node.currentTerm.Id())
	}
	r.node.election.Reset()
	if req.PrevLogIndex==0&&req.PrevLogTerm==0&&len(req.Entries)==0{
		res.Success=true
		return nil
	}else if req.PrevLogIndex>0{
		if req.PrevLogIndex>r.node.lastLogIndex{
			res.Success=false
			return nil
		}else if  r.node.log.lookup(req.PrevLogIndex)==nil{
			res.Success=false
			return nil
		}else if r.node.log.lookup(req.PrevLogIndex).Term!=req.PrevLogTerm{
			r.node.log.deleteAfter(req.PrevLogIndex)
			r.node.nextIndex=r.node.lastLogIndex+1
			res.NextIndex=r.node.nextIndex
			//r.node.commitIndex=0
			res.Success=false
			return nil
		}
	}
	if req.LeaderCommit>r.node.commitIndex.Id() {
		//var commitIndex=r.node.commitIndex
		r.node.commitIndex.Set(minUint64(req.LeaderCommit,r.node.lastLogIndex))
		//if r.node.commitIndex>commitIndex{
		//	Tracef("raft.HandleAppendEntries %s commitIndex %d==>%d",r.node.address, commitIndex,r.node.commitIndex)
		//}
	}
	if len(req.Entries)>0&&req.PrevLogIndex==r.node.lastLogIndex{
		if req.PrevLogIndex+1==req.Entries[0].Index{
			res.Success=r.node.log.appendEntries(req.Entries)
			r.node.nextIndex=r.node.lastLogIndex+1
			res.NextIndex=r.node.nextIndex
			return nil
		}
	}
	res.Success=false
	return nil
}
func (r *raft) HandleInstallSnapshot(req *InstallSnapshotRequest,res *InstallSnapshotResponse)error {
	return nil
}



