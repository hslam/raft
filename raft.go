package raft

import (
	"github.com/hslam/rpc"
	"sync"
	"time"
)

type Raft interface {
	RequestVote(addr string) (ok bool)
	AppendEntries(addr string, prevLogIndex, prevLogTerm uint64, entries []*Entry) (nextIndex, term uint64, success, ok bool)
	InstallSnapshot(addr string, LastIncludedIndex, LastIncludedTerm, Offset uint64, Data []byte, Done bool) (offset, nextIndex uint64, ok bool)
	HandleRequestVote(req *RequestVoteRequest, res *RequestVoteResponse) error
	HandleAppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse) error
	HandleInstallSnapshot(req *InstallSnapshotRequest, res *InstallSnapshotResponse) error
}

type raft struct {
	node                   *Node
	callPool               *sync.Pool
	donePool               *sync.Pool
	hearbeatTimeout        time.Duration
	requestVoteTimeout     time.Duration
	appendEntriesTimeout   time.Duration
	installSnapshotTimeout time.Duration
}

func newRaft(node *Node) Raft {
	return &raft{
		node:                   node,
		callPool:               &sync.Pool{New: func() interface{} { return &rpc.Call{} }},
		donePool:               &sync.Pool{New: func() interface{} { return make(chan *rpc.Call, 10) }},
		hearbeatTimeout:        DefaultHearbeatTimeout,
		requestVoteTimeout:     DefaultRequestVoteTimeout,
		appendEntriesTimeout:   DefaultAppendEntriesTimeout,
		installSnapshotTimeout: DefaultInstallSnapshotTimeout,
	}
}

func (r *raft) RequestVote(addr string) (ok bool) {
	var req = &RequestVoteRequest{}
	req.Term = r.node.currentTerm.Id()
	req.CandidateId = r.node.address
	req.LastLogIndex = r.node.lastLogIndex
	req.LastLogTerm = r.node.lastLogTerm
	done := r.donePool.Get().(chan *rpc.Call)
	call := r.callPool.Get().(*rpc.Call)
	call.ServiceMethod = r.node.rpcs.RequestVoteServiceName()
	call.Args = req
	call.Reply = &RequestVoteResponse{}
	call.Done = done
	r.node.rpcs.RoundTrip(addr, call)
	select {
	case call := <-done:
		for len(done) > 0 {
			<-done
		}
		r.donePool.Put(done)
		if call.Error != nil {
			Tracef("raft.RequestVote %s recv %s vote error %s", r.node.address, addr, call.Error.Error())
			*call = rpc.Call{}
			r.callPool.Put(call)
			return false
		}
		res := call.Reply.(*RequestVoteResponse)
		*call = rpc.Call{}
		r.callPool.Put(call)
		if res.Term > r.node.currentTerm.Id() {
			r.node.currentTerm.Set(res.Term)
			r.node.stepDown()
		}
		//Tracef("raft.RequestVote %s recv %s vote %t",r.node.address,addr,res.VoteGranted)
		if res.VoteGranted {
			r.node.votes.vote <- newVote(addr, req.Term, 1)
		} else {
			r.node.votes.vote <- newVote(addr, req.Term, 0)
		}
		return true
	case <-time.After(r.requestVoteTimeout):
		Tracef("raft.RequestVote %s recv %s vote time out", r.node.address, addr)
		r.node.votes.vote <- newVote(addr, req.Term, 0)
	}
	return false
}

func (r *raft) AppendEntries(addr string, prevLogIndex, prevLogTerm uint64, entries []*Entry) (nextIndex uint64, term uint64, success bool, ok bool) {
	var req = &AppendEntriesRequest{}
	req.Term = r.node.currentTerm.Id()
	req.LeaderId = r.node.leader
	req.LeaderCommit = r.node.commitIndex
	req.PrevLogIndex = prevLogIndex
	req.PrevLogTerm = prevLogTerm
	req.Entries = entries
	var timeout = r.appendEntriesTimeout
	if len(entries) == 0 {
		timeout = r.hearbeatTimeout
	}
	done := r.donePool.Get().(chan *rpc.Call)
	call := r.callPool.Get().(*rpc.Call)
	call.ServiceMethod = r.node.rpcs.AppendEntriesServiceName()
	call.Args = req
	call.Reply = &AppendEntriesResponse{}
	call.Done = done
	r.node.rpcs.RoundTrip(addr, call)
	select {
	case call := <-done:
		for len(done) > 0 {
			<-done
		}
		r.donePool.Put(done)
		if call.Error != nil {
			Tracef("raft.AppendEntries %s -> %s error %s", r.node.address, addr, call.Error.Error())
			*call = rpc.Call{}
			r.callPool.Put(call)
			return 0, 0, false, false
		}
		res := call.Reply.(*AppendEntriesResponse)
		*call = rpc.Call{}
		r.callPool.Put(call)
		if res.Term > r.node.currentTerm.Id() {
			r.node.currentTerm.Set(res.Term)
			r.node.stepDown()
			if len(entries) > 0 {
				return res.NextIndex, res.Term, false, true
			}
		}
		//Tracef("raft.AppendEntries %s -> %s",r.node.address,addr)
		return res.NextIndex, res.Term, res.Success, true
	case <-time.After(timeout):
		Tracef("raft.AppendEntries %s -> %s time out", r.node.address, addr)
	}
	return 0, 0, false, false
}

func (r *raft) InstallSnapshot(addr string, LastIncludedIndex, LastIncludedTerm, Offset uint64, Data []byte, Done bool) (offset uint64, nextIndex uint64, ok bool) {
	var req = &InstallSnapshotRequest{}
	req.Term = r.node.currentTerm.Id()
	req.LeaderId = r.node.leader
	req.LastIncludedIndex = LastIncludedIndex
	req.LastIncludedTerm = LastIncludedTerm
	req.Offset = Offset
	req.Data = Data
	req.Done = Done
	done := r.donePool.Get().(chan *rpc.Call)
	call := r.callPool.Get().(*rpc.Call)
	call.ServiceMethod = r.node.rpcs.InstallSnapshotServiceName()
	call.Args = req
	call.Reply = &InstallSnapshotResponse{}
	call.Done = done
	r.node.rpcs.RoundTrip(addr, call)
	select {
	case call := <-done:
		for len(done) > 0 {
			<-done
		}
		r.donePool.Put(done)
		if call.Error != nil {
			Tracef("raft.InstallSnapshot %s -> %s error %s", r.node.address, addr, call.Error.Error())
			*call = rpc.Call{}
			r.callPool.Put(call)
			return 0, 0, false
		}
		res := call.Reply.(*InstallSnapshotResponse)
		*call = rpc.Call{}
		r.callPool.Put(call)
		if res.Term > r.node.currentTerm.Id() {
			r.node.currentTerm.Set(res.Term)
			r.node.stepDown()
		}
		//Tracef("raft.InstallSnapshot %s -> %s offset %d",r.node.address,addr,res.Offset)
		return res.Offset, res.NextIndex, true
	case <-time.After(r.installSnapshotTimeout):
		Tracef("raft.InstallSnapshot %s -> %s time out", r.node.address, addr)
	}
	return 0, 0, false
}

func (r *raft) HandleRequestVote(req *RequestVoteRequest, res *RequestVoteResponse) error {
	res.Term = r.node.currentTerm.Id()
	if req.Term < r.node.currentTerm.Id() {
		res.VoteGranted = false
		return nil
	} else if req.Term > r.node.currentTerm.Id() && req.LastLogIndex >= r.node.lastLogIndex && req.LastLogTerm >= r.node.lastLogTerm {
		r.node.currentTerm.Set(req.Term)
		r.node.votedFor.Reset()
		r.node.stepDown()
	}
	if r.node.state.String() == Leader || r.node.state.String() == Candidate || r.node.leader != "" {
		res.VoteGranted = false
		return nil
	} else if (r.node.votedFor.String() == "" || r.node.votedFor.String() == req.CandidateId) && req.LastLogIndex >= r.node.lastLogIndex && req.LastLogTerm >= r.node.lastLogTerm {
		res.VoteGranted = true
		r.node.votedFor.Set(req.CandidateId)
		r.node.stepDown()
	}
	return nil

}

func (r *raft) HandleAppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse) error {
	res.Term = r.node.currentTerm.Id()
	res.NextIndex = r.node.nextIndex
	if req.Term < r.node.currentTerm.Id() {
		res.Success = false
		return nil
	} else if req.Term > r.node.currentTerm.Id() {
		r.node.currentTerm.Set(req.Term)
		r.node.stepDown()
		if r.node.state.String() == Leader || r.node.state.String() == Candidate {
			res.Success = false
			return nil
		}
	}
	if r.node.leader == "" {
		r.node.votedFor.Set(req.LeaderId)
		r.node.leader = req.LeaderId
		Tracef("raft.HandleAppendEntries %s State:%s leader-%s Term:%d", r.node.address, r.node.State(), r.node.leader, r.node.currentTerm.Id())
	}
	if r.node.leader != req.LeaderId {
		res.Success = false
		r.node.stepDown()
		return nil
	}
	if r.node.state.String() == Leader || r.node.state.String() == Candidate {
		r.node.stepDown()
		Tracef("raft.HandleAppendEntries %s State:%s Term:%d", r.node.address, r.node.State(), r.node.currentTerm.Id())
	}
	r.node.election.Reset()
	if req.PrevLogIndex == 0 && req.PrevLogTerm == 0 && len(req.Entries) == 0 {
		res.Success = true
		return nil
	} else if req.PrevLogIndex > 0 {
		if ok := r.node.log.consistencyCheck(req.PrevLogIndex, req.PrevLogTerm); !ok {
			res.NextIndex = r.node.nextIndex
			res.Success = false
			return nil
		}
	}

	if req.LeaderCommit > r.node.commitIndex {
		//var commitIndex=r.node.commitIndex
		r.node.commitIndex = minUint64(req.LeaderCommit, r.node.lastLogIndex)
		//if r.node.commitIndex>commitIndex{
		//	Tracef("raft.HandleAppendEntries %s commitIndex %d==>%d",r.node.address, commitIndex,r.node.commitIndex)
		//}
	}
	if len(req.Entries) > 0 && req.PrevLogIndex == r.node.lastLogIndex {
		if req.PrevLogIndex+1 == req.Entries[0].Index {
			res.Success = r.node.log.appendEntries(req.Entries)
			r.node.nextIndex = r.node.lastLogIndex + 1
			res.NextIndex = r.node.nextIndex
			return nil
		}
	}
	res.Success = false
	return nil
}

func (r *raft) HandleInstallSnapshot(req *InstallSnapshotRequest, res *InstallSnapshotResponse) error {
	res.Term = r.node.currentTerm.Id()
	if req.Term < r.node.currentTerm.Id() {
		return nil
	} else if req.Term > r.node.currentTerm.Id() {
		r.node.currentTerm.Set(req.Term)
		r.node.stepDown()
	}
	Tracef("raft.HandleInstallSnapshot offset %d len %d done %t", req.Offset, len(req.Data), req.Done)
	if req.Offset == 0 {
		r.node.stateMachine.snapshotReadWriter.clear()
		r.node.stateMachine.snapshotReadWriter.done = false
	}
	r.node.stateMachine.append(req.Offset, req.Data)
	if req.Done {
		if r.node.leader != req.LeaderId {
			r.node.leader = req.LeaderId
		}
		r.node.stateMachine.snapshotReadWriter.done = true
		r.node.stateMachine.snapshotReadWriter.lastIncludedIndex.Set(req.LastIncludedIndex)
		r.node.stateMachine.snapshotReadWriter.lastIncludedTerm.Set(req.LastIncludedTerm)
		r.node.stateMachine.snapshotReadWriter.untar()
		r.node.recover()
	}
	offset, err := r.node.storage.Size(r.node.stateMachine.snapshotReadWriter.FileName())
	if err != nil {
		return nil
	}
	res.Offset = uint64(offset)
	res.NextIndex = r.node.nextIndex
	return nil
}
