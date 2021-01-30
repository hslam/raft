// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"context"
	"time"
)

// Raft represents the raft service.
type Raft interface {
	CallRequestVote(addr string) (ok bool)
	CallAppendEntries(addr string, prevLogIndex, prevLogTerm uint64, entries []*Entry) (nextIndex, term uint64, success, ok bool)
	CallInstallSnapshot(addr string, LastIncludedIndex, LastIncludedTerm, Offset uint64, Data []byte, Done bool) (offset, nextIndex uint64, ok bool)
	RequestVote(req *RequestVoteRequest, res *RequestVoteResponse) error
	AppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse) error
	InstallSnapshot(req *InstallSnapshotRequest, res *InstallSnapshotResponse) error
}

type raft struct {
	node                   *node
	hearbeatTimeout        time.Duration
	requestVoteTimeout     time.Duration
	appendEntriesTimeout   time.Duration
	installSnapshotTimeout time.Duration
}

func newRaft(n *node) Raft {
	return &raft{
		node:                   n,
		hearbeatTimeout:        defaultHearbeatTimeout,
		requestVoteTimeout:     defaultRequestVoteTimeout,
		appendEntriesTimeout:   defaultAppendEntriesTimeout,
		installSnapshotTimeout: defaultInstallSnapshotTimeout,
	}
}

func (r *raft) CallRequestVote(addr string) (ok bool) {
	var req = &RequestVoteRequest{}
	req.Term = r.node.currentTerm.Load()
	req.CandidateID = r.node.address
	req.LastLogIndex = r.node.lastLogIndex
	req.LastLogTerm = r.node.lastLogTerm
	var res = &RequestVoteResponse{}
	ctx, cancel := context.WithTimeout(context.Background(), r.requestVoteTimeout)
	defer cancel()
	err := r.node.rpcs.CallWithContext(ctx, addr, r.node.rpcs.RequestVoteServiceName(), req, res)
	if err != nil {
		r.node.logger.Tracef("raft.CallRequestVote %s recv %s vote error %s", r.node.address, addr, err.Error())
		return false
	}
	if res.Term > r.node.currentTerm.Load() {
		r.node.currentTerm.Store(res.Term)
		r.node.stepDown(false)
	}
	if res.VoteGranted {
		r.node.votes.vote <- newVote(addr, req.Term, 1)
	} else {
		r.node.votes.vote <- newVote(addr, req.Term, 0)
	}
	return true
}

func (r *raft) CallAppendEntries(addr string, prevLogIndex, prevLogTerm uint64, entries []*Entry) (nextIndex uint64, term uint64, success bool, ok bool) {
	var req = &AppendEntriesRequest{}
	req.Term = r.node.currentTerm.Load()
	req.LeaderID = r.node.leader.Load()
	req.LeaderCommit = r.node.commitIndex.ID()
	req.PrevLogIndex = prevLogIndex
	req.PrevLogTerm = prevLogTerm
	req.Entries = entries
	var timeout = r.appendEntriesTimeout
	if len(entries) == 0 {
		timeout = r.hearbeatTimeout
	}
	var res = &AppendEntriesResponse{}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := r.node.rpcs.CallWithContext(ctx, addr, r.node.rpcs.AppendEntriesServiceName(), req, res)
	if err != nil {
		//r.node.logger.Tracef("raft.CallAppendEntries %s -> %s error %s", r.node.address, addr, err.Error())
		return 0, 0, false, false
	}
	if res.Term > r.node.currentTerm.Load() {
		r.node.currentTerm.Store(res.Term)
		r.node.stepDown(false)
		if len(entries) > 0 {
			return res.NextIndex, res.Term, false, true
		}
	}
	return res.NextIndex, res.Term, res.Success, true
}

func (r *raft) CallInstallSnapshot(addr string, LastIncludedIndex, LastIncludedTerm, Offset uint64, Data []byte, Done bool) (offset uint64, nextIndex uint64, ok bool) {
	var req = &InstallSnapshotRequest{}
	req.Term = r.node.currentTerm.Load()
	req.LeaderID = r.node.leader.Load()
	req.LastIncludedIndex = LastIncludedIndex
	req.LastIncludedTerm = LastIncludedTerm
	req.Offset = Offset
	req.Data = Data
	req.Done = Done
	var res = &InstallSnapshotResponse{}
	ctx, cancel := context.WithTimeout(context.Background(), r.installSnapshotTimeout)
	defer cancel()
	err := r.node.rpcs.CallWithContext(ctx, addr, r.node.rpcs.InstallSnapshotServiceName(), req, res)
	if err != nil {
		r.node.logger.Tracef("raft.CallInstallSnapshot %s -> %s error %s", r.node.address, addr, err.Error())
		return 0, 0, false
	}
	if res.Term > r.node.currentTerm.Load() {
		r.node.currentTerm.Store(res.Term)
		r.node.stepDown(false)
	}
	return res.Offset, res.NextIndex, true
}

func (r *raft) RequestVote(req *RequestVoteRequest, res *RequestVoteResponse) error {
	res.Term = r.node.currentTerm.Load()
	if req.Term < r.node.currentTerm.Load() {
		res.VoteGranted = false
		return nil
	} else if req.Term > r.node.currentTerm.Load() {
		r.node.currentTerm.Store(req.Term)
		r.node.votedFor.Store("")
		if r.node.state.String() != follower {
			r.node.stepDown(true)
		}
	}
	isUpToDate := req.LastLogIndex >= r.node.lastLogIndex && req.LastLogTerm >= r.node.lastLogTerm
	if (r.node.votedFor.Load() == "" || r.node.votedFor.Load() == req.CandidateID) && isUpToDate {
		res.VoteGranted = true
		r.node.votedFor.Store(cloneString(req.CandidateID))
		r.node.stepDown(true)
	}
	return nil

}

func (r *raft) AppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse) error {
	res.Term = r.node.currentTerm.Load()
	res.NextIndex = r.node.nextIndex
	if req.Term < r.node.currentTerm.Load() {
		res.Success = false
		return nil
	} else if req.Term > r.node.currentTerm.Load() {
		r.node.currentTerm.Store(req.Term)
		if r.node.state.String() != follower {
			r.node.stepDown(true)
		}
	}
	if r.node.leader.Load() == "" {
		r.node.votedFor.Store(cloneString(req.LeaderID))
		r.node.leader.Store(cloneString(req.LeaderID))
		r.node.logger.Tracef("raft.HandleAppendEntries %s State:%s leader-%s Term:%d", r.node.address, r.node.State(), r.node.leader.Load(), r.node.currentTerm.Load())
		if r.node.leaderChange != nil {
			go r.node.leaderChange()
		}
	}
	if r.node.leader.Load() != req.LeaderID {
		res.Success = false
		r.node.stepDown(true)
		return nil
	}
	if r.node.state.String() == candidate {
		r.node.stepDown(true)
		r.node.logger.Tracef("raft.HandleAppendEntries %s State:%s Term:%d", r.node.address, r.node.State(), r.node.currentTerm.Load())
	}
	r.node.election.Reset()

	if req.PrevLogIndex > 0 {
		if ok := r.node.log.consistencyCheck(req.PrevLogIndex, req.PrevLogTerm); !ok {
			res.NextIndex = r.node.nextIndex
			res.Success = false
			return nil
		}
		r.node.ready = true
	}

	if req.LeaderCommit > r.node.commitIndex.ID() {
		r.node.commitIndex.Set(minUint64(req.LeaderCommit, r.node.lastLogIndex))
	}
	if len(req.Entries) == 0 {
		res.Success = true
		return nil
	}
	if req.PrevLogIndex+1 == req.Entries[0].Index {
		res.Success = r.node.log.appendEntries(req.Entries)
		r.node.nextIndex = r.node.lastLogIndex + 1
		res.NextIndex = r.node.nextIndex
		return nil
	}
	res.Success = false
	return nil
}

func (r *raft) InstallSnapshot(req *InstallSnapshotRequest, res *InstallSnapshotResponse) error {
	res.Term = r.node.currentTerm.Load()
	if req.Term < r.node.currentTerm.Load() {
		return nil
	} else if req.Term > r.node.currentTerm.Load() {
		r.node.currentTerm.Store(req.Term)
		if r.node.state.String() != follower {
			r.node.stepDown(true)
		}
	}
	r.node.logger.Tracef("raft.HandleInstallSnapshot offset %d len %d done %t", req.Offset, len(req.Data), req.Done)
	if req.Offset == 0 {
		r.node.stateMachine.snapshotReadWriter.finish.Store(false)
		r.node.stateMachine.snapshotReadWriter.clear()
	}
	r.node.stateMachine.append(req.Offset, req.Data)
	offset, err := r.node.storage.Size(r.node.stateMachine.snapshotReadWriter.FileName())
	if err != nil {
		return nil
	}
	res.Offset = uint64(offset)
	if req.Done {
		if r.node.leader.Load() == req.LeaderID {
			r.node.reset()
			r.node.stateMachine.snapshotReadWriter.lastIncludedIndex.Set(req.LastIncludedIndex)
			r.node.stateMachine.snapshotReadWriter.lastIncludedTerm.Set(req.LastIncludedTerm)
			r.node.stateMachine.snapshotReadWriter.finish.Store(true)
			if err := r.node.stateMachine.snapshotReadWriter.untar(); err == nil {
				r.node.recover()
			}
		} else {
			r.node.leader.Store(cloneString(req.LeaderID))
			r.node.stateMachine.snapshotReadWriter.finish.Store(false)
			r.node.stateMachine.snapshotReadWriter.clear()
		}
	}
	res.NextIndex = r.node.nextIndex
	return nil
}
