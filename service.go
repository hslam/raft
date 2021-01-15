// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

type service struct {
	node *node
}

//RPCMethodName
func (s *service) RequestVote(req *RequestVoteRequest, res *RequestVoteResponse) error {
	return s.node.raft.RequestVote(req, res)
}
func (s *service) AppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse) error {
	return s.node.raft.AppendEntries(req, res)
}
func (s *service) InstallSnapshot(req *InstallSnapshotRequest, res *InstallSnapshotResponse) error {
	return s.node.raft.InstallSnapshot(req, res)
}
func (s *service) GetLeader(req *GetLeaderRequest, res *GetLeaderResponse) error {
	return s.node.cluster.GetLeader(req, res)
}
func (s *service) AddMember(req *AddMemberRequest, res *AddMemberResponse) error {
	return s.node.cluster.AddMember(req, res)
}
func (s *service) RemoveMember(req *RemoveMemberRequest, res *RemoveMemberResponse) error {
	return s.node.cluster.RemoveMember(req, res)
}
func (s *service) SetMeta(req *SetMetaRequest, res *SetMetaResponse) error {
	return s.node.cluster.SetMeta(req, res)
}
func (s *service) GetMeta(req *GetMetaRequest, res *GetMetaResponse) error {
	return s.node.cluster.GetMeta(req, res)
}

//RPCShortMethodName
func (s *service) R(req *RequestVoteRequest, res *RequestVoteResponse) error {
	return s.RequestVote(req, res)
}
func (s *service) A(req *AppendEntriesRequest, res *AppendEntriesResponse) error {
	return s.AppendEntries(req, res)
}
func (s *service) I(req *InstallSnapshotRequest, res *InstallSnapshotResponse) error {
	return s.InstallSnapshot(req, res)
}
func (s *service) Q(req *GetLeaderRequest, res *GetLeaderResponse) error {
	return s.GetLeader(req, res)
}
func (s *service) J(req *AddMemberRequest, res *AddMemberResponse) error {
	return s.AddMember(req, res)
}
func (s *service) L(req *RemoveMemberRequest, res *RemoveMemberResponse) error {
	return s.RemoveMember(req, res)
}
func (s *service) S(req *SetMetaRequest, res *SetMetaResponse) error {
	return s.SetMeta(req, res)
}
func (s *service) G(req *GetMetaRequest, res *GetMetaResponse) error {
	return s.GetMeta(req, res)
}
