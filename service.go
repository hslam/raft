// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

// Service represents the RPCs service.
type Service interface {
	RequestVote(req *RequestVoteRequest, res *RequestVoteResponse) error
	AppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse) error
	InstallSnapshot(req *InstallSnapshotRequest, res *InstallSnapshotResponse) error
	GetLeader(req *GetLeaderRequest, res *GetLeaderResponse) error
	AddMember(req *AddMemberRequest, res *AddMemberResponse) error
	RemoveMember(req *RemoveMemberRequest, res *RemoveMemberResponse) error
	SetMeta(req *SetMetaRequest, res *SetMetaResponse) error
	GetMeta(req *GetMetaRequest, res *GetMetaResponse) error
}

type service struct {
	raft    *raft
	cluster *cluster
}

func newService(n *node) *service {
	return &service{
		raft:    newRaft(n),
		cluster: newCluster(n),
	}
}

func (s *service) RequestVote(req *RequestVoteRequest, res *RequestVoteResponse) error {
	return s.raft.RequestVote(req, res)
}

func (s *service) AppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse) error {
	return s.raft.AppendEntries(req, res)
}

func (s *service) InstallSnapshot(req *InstallSnapshotRequest, res *InstallSnapshotResponse) error {
	return s.raft.InstallSnapshot(req, res)
}

func (s *service) GetLeader(req *GetLeaderRequest, res *GetLeaderResponse) error {
	return s.cluster.GetLeader(req, res)
}

func (s *service) AddMember(req *AddMemberRequest, res *AddMemberResponse) error {
	return s.cluster.AddMember(req, res)
}

func (s *service) RemoveMember(req *RemoveMemberRequest, res *RemoveMemberResponse) error {
	return s.cluster.RemoveMember(req, res)
}

func (s *service) SetMeta(req *SetMetaRequest, res *SetMetaResponse) error {
	return s.cluster.SetMeta(req, res)
}

func (s *service) GetMeta(req *GetMetaRequest, res *GetMetaResponse) error {
	return s.cluster.GetMeta(req, res)
}
