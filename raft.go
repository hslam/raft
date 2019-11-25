package raft

import (
	"time"
)

type Raft interface {
	RequestVote(addr string) (ok bool)
	AppendEntries(addr string,prevLogIndex,prevLogTerm uint64,entries []*Entry)  (nextIndex,term uint64,success ,ok bool)
	InstallSnapshot(addr string,LastIncludedIndex,LastIncludedTerm,Offset uint64,Data []byte,Done bool) (offset,nextIndex uint64,ok bool)
	HandleRequestVote(req *RequestVoteRequest, res *RequestVoteResponse)error
	HandleAppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse)error
	HandleInstallSnapshot(req *InstallSnapshotRequest,res *InstallSnapshotResponse)error
	QueryLeader(addr string)(term uint64,leaderId string,ok bool)
	AddPeer(addr string,info *NodeInfo)(success bool,ok bool)
	RemovePeer(addr string,Address string)(success bool,ok bool)
	HandleQueryLeader(req *QueryLeaderRequest,res *QueryLeaderResponse)error
	HandleAddPeer(req *AddPeerRequest,res *AddPeerResponse)error
	HandleRemovePeer(req *RemovePeerRequest,res *RemovePeerResponse)error
}

type raft struct {
	node	*Node
	hearbeatTimeout			time.Duration
	requestVoteTimeout		time.Duration
	appendEntriesTimeout	time.Duration
	installSnapshotTimeout	time.Duration
	queryLeaderTimeout		time.Duration
	addPeerTimeout			time.Duration
	removePeerTimeout		time.Duration
}

func newRaft(node	*Node) Raft{
	return &raft{
		node:node,
		hearbeatTimeout:		DefaultHearbeatTimeout,
		requestVoteTimeout:		DefaultRequestVoteTimeout,
		appendEntriesTimeout:	DefaultAppendEntriesTimeout,
		installSnapshotTimeout:	DefaultInstallSnapshotTimeout,
		queryLeaderTimeout: 	DefaultQueryLeaderTimeout,
		addPeerTimeout:			DefaultAddPeerTimeout,
		removePeerTimeout:		DefaultRemovePeerTimeout,
	}
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
		if err := rpcs.Call(rpcs.RequestVoteServiceName(),req, res,addr); err != nil {
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
	var timeout=r.appendEntriesTimeout
	if len(entries)==0{
		timeout=r.hearbeatTimeout
	}
	var ch =make(chan *AppendEntriesResponse,1)
	var errCh =make(chan error,1)
	go func (rpcs *RPCs,addr string,req *AppendEntriesRequest, ch chan *AppendEntriesResponse,errCh chan error) {
		var res =&AppendEntriesResponse{}
		if err := rpcs.Call(rpcs.AppendEntriesServiceName(),req, res,addr); err != nil {
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
	case <-time.After(timeout):
		Tracef("raft.AppendEntries %s -> %s time out",r.node.address,addr)
	}
	return 0,0,false,false
}

func (r *raft) InstallSnapshot(addr string,LastIncludedIndex,LastIncludedTerm,Offset uint64,Data []byte,Done bool)(offset uint64,nextIndex uint64,ok bool){
	var req =&InstallSnapshotRequest{}
	req.Term=r.node.currentTerm.Id()
	req.LeaderId=r.node.leader
	req.LastIncludedIndex=LastIncludedIndex
	req.LastIncludedTerm=LastIncludedTerm
	req.Offset=Offset
	req.Data=Data
	req.Done=Done
	var ch =make(chan *InstallSnapshotResponse,1)
	var errCh =make(chan error,1)
	go func (rpcs *RPCs,addr string,req *InstallSnapshotRequest, ch chan *InstallSnapshotResponse,errCh chan error) {
		var res =&InstallSnapshotResponse{}
		if err := rpcs.Call(rpcs.InstallSnapshotServiceName(),req, res,addr); err != nil {
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
		//Tracef("raft.InstallSnapshot %s -> %s offset %d",r.node.address,addr,res.Offset)
		return res.Offset,res.NextIndex,true
	case err:=<-errCh:
		Tracef("raft.InstallSnapshot %s -> %s error %s",r.node.address,addr,err.Error())
		return 0,0,false
	case <-time.After(r.installSnapshotTimeout):
		Tracef("raft.InstallSnapshot %s -> %s time out",r.node.address,addr)
	}
	return 0,0,false
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
	if r.node.state.String()==Leader||r.node.state.String()==Candidate||r.node.leader!=""{
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
		if ok:=r.node.log.consistencyCheck(req.PrevLogIndex,req.PrevLogTerm);!ok{
			res.NextIndex=r.node.nextIndex
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
	res.Term=r.node.currentTerm.Id()
	if req.Term<r.node.currentTerm.Id(){
		return nil
	}else if req.Term>r.node.currentTerm.Id(){
		r.node.currentTerm.Set(req.Term)
		r.node.stepDown()
	}
	Tracef("raft.HandleInstallSnapshot offset %d len %d done %t", req.Offset,len(req.Data),req.Done)
	if req.Offset==0{
		r.node.stateMachine.snapshotReadWriter.clear()
		r.node.stateMachine.snapshotReadWriter.done=false
	}
	r.node.stateMachine.append(req.Offset,req.Data)
	if req.Done{
		if r.node.leader!=req.LeaderId{
			r.node.leader=req.LeaderId
		}
		r.node.stateMachine.snapshotReadWriter.done=true
		r.node.stateMachine.snapshotReadWriter.lastIncludedIndex.Set(req.LastIncludedIndex)
		r.node.stateMachine.snapshotReadWriter.lastIncludedTerm.Set(req.LastIncludedTerm)
		r.node.stateMachine.snapshotReadWriter.untar()
		r.node.recover()
	}
	offset,err:=r.node.storage.Size(r.node.stateMachine.snapshotReadWriter.FileName())
	if err!=nil{
		return nil
	}
	res.Offset=uint64(offset)
	res.NextIndex=r.node.nextIndex
	return nil
}

func (r *raft) QueryLeader(addr string)(term uint64,leaderId string,ok bool){
	var req =&QueryLeaderRequest{}
	var ch =make(chan *QueryLeaderResponse,1)
	var errCh =make(chan error,1)
	go func (rpcs *RPCs,addr string,req *QueryLeaderRequest, ch chan *QueryLeaderResponse,errCh chan error) {
		var res =&QueryLeaderResponse{}
		if err := rpcs.Call(rpcs.QueryLeaderServiceName(),req, res,addr); err != nil {
			errCh<-err
		}else {
			ch <- res
		}
	}(r.node.rpcs,addr,req,ch,errCh)
	select {
	case res:=<-ch:
		Tracef("raft.QueryLeader %s -> %s LeaderId %s",r.node.address,addr,res.LeaderId)
		return res.Term,res.LeaderId,true
	case err:=<-errCh:
		Tracef("raft.QueryLeader %s -> %s error %s",r.node.address,addr,err.Error())
		return 0,"",false
	case <-time.After(r.addPeerTimeout):
		Tracef("raft.QueryLeader %s -> %s time out",r.node.address,addr)
	}
	return 0,"",false
}
func (r *raft) AddPeer(addr string,info *NodeInfo)(success bool,ok bool){
	var req =&AddPeerRequest{}
	req.Node=info
	var ch =make(chan *AddPeerResponse,1)
	var errCh =make(chan error,1)
	go func (rpcs *RPCs,addr string,req *AddPeerRequest, ch chan *AddPeerResponse,errCh chan error) {
		var res =&AddPeerResponse{}
		if err := rpcs.Call(rpcs.AddPeerServiceName(),req, res,addr); err != nil {
			errCh<-err
		}else {
			ch <- res
		}
	}(r.node.rpcs,addr,req,ch,errCh)
	select {
	case res:=<-ch:
		Tracef("raft.AddPeer %s -> %s Success %t",r.node.address,addr,res.Success)
		return res.Success,true
	case err:=<-errCh:
		Tracef("raft.AddPeer %s -> %s error %s",r.node.address,addr,err.Error())
		return false,false
	case <-time.After(r.addPeerTimeout):
		Tracef("raft.AddPeer %s -> %s time out",r.node.address,addr)
	}
	return false,false
}
func (r *raft) RemovePeer(addr string,Address string)(success bool,ok bool){
	var req =&RemovePeerRequest{}
	req.Address=Address
	var ch =make(chan *RemovePeerResponse,1)
	var errCh =make(chan error,1)
	go func (rpcs *RPCs,addr string,req *RemovePeerRequest, ch chan *RemovePeerResponse,errCh chan error) {
		var res =&RemovePeerResponse{}
		if err := rpcs.Call(rpcs.RemovePeerServiceName(),req, res,addr); err != nil {
			errCh<-err
		}else {
			ch <- res
		}
	}(r.node.rpcs,addr,req,ch,errCh)
	select {
	case res:=<-ch:
		Tracef("raft.RemovePeer %s -> %s Success %t",r.node.address,addr,res.Success)
		return res.Success,true
	case err:=<-errCh:
		Tracef("raft.RemovePeer %s -> %s error %s",r.node.address,addr,err.Error())
		return false,false
	case <-time.After(r.removePeerTimeout):
		Tracef("raft.RemovePeer %s -> %s time out",r.node.address,addr)
	}
	return false,false
}
func (r *raft) HandleQueryLeader(req *QueryLeaderRequest,res *QueryLeaderResponse)error {
	if r.node.leader!=""{
		res.LeaderId=r.node.leader
		res.Term=r.node.term()
		return nil
	}
	peers:=r.node.Peers()
	for i:=0;i<len(peers) ;i++  {
		term,leaderId,ok:=r.QueryLeader(peers[i])
		if ok{
			res.LeaderId=leaderId
			res.Term=term
			return nil
		}
	}
	return ErrNotLeader
}
func (r *raft) HandleAddPeer(req *AddPeerRequest,res *AddPeerResponse)error {
	if r.node.IsLeader(){
		_, err := r.node.do(NewAddPeerCommand(req.Node),DefaultCommandTimeout)
		if err==nil{
			_, err = r.node.do(NewReconfigurationCommand(),DefaultCommandTimeout)
			if err==nil{
				res.Success=true
				return nil
			}
			return err
		}
		return err
	}else {
		leader:=r.node.Leader()
		if leader!=""{
			return r.node.rpcs.Call(r.node.rpcs.AddPeerServiceName(),req,res,leader)
		}
		peers:=r.node.Peers()
		for i:=0;i<len(peers) ;i++  {
			_,leaderId,ok:=r.QueryLeader(peers[i])
			if leaderId!=""&&ok{
				return r.node.rpcs.Call(r.node.rpcs.AddPeerServiceName(),req,res,leaderId)
			}
		}
		return ErrNotLeader
	}
	return nil
}

func (r *raft) HandleRemovePeer(req *RemovePeerRequest,res *RemovePeerResponse)error {
	if r.node.IsLeader(){
		_, err := r.node.do(NewRemovePeerCommand(req.Address),DefaultCommandTimeout)
		if err==nil{
			_, err = r.node.do(NewReconfigurationCommand(),DefaultCommandTimeout)
			if err==nil{
				res.Success=true
				return nil
			}
			return err
		}
		return err
	}else {
		leader:=r.node.Leader()
		if leader!=""{
			return r.node.rpcs.Call(r.node.rpcs.RemovePeerServiceName(),req,res,leader)
		}
		peers:=r.node.Peers()
		for i:=0;i<len(peers) ;i++  {
			_,leaderId,ok:=r.QueryLeader(peers[i])
			if leaderId!=""&&ok{
				return r.node.rpcs.Call(r.node.rpcs.RemovePeerServiceName(),req,res,leaderId)
			}
		}
		return ErrNotLeader
	}
	return nil
}



