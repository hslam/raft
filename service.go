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
func (s *service) QueryLeader(req *QueryLeaderRequest, res *QueryLeaderResponse) error {
	return s.node.cluster.QueryLeader(req, res)
}
func (s *service) SetPeer(req *SetPeerRequest, res *SetPeerResponse) error {
	return s.node.cluster.SetPeer(req, res)
}
func (s *service) RemovePeer(req *RemovePeerRequest, res *RemovePeerResponse) error {
	return s.node.cluster.RemovePeer(req, res)
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
func (s *service) Q(req *QueryLeaderRequest, res *QueryLeaderResponse) error {
	return s.QueryLeader(req, res)
}
func (s *service) J(req *SetPeerRequest, res *SetPeerResponse) error {
	return s.SetPeer(req, res)
}
func (s *service) L(req *RemovePeerRequest, res *RemovePeerResponse) error {
	return s.RemovePeer(req, res)
}
func (s *service) S(req *SetMetaRequest, res *SetMetaResponse) error {
	return s.SetMeta(req, res)
}
func (s *service) G(req *GetMetaRequest, res *GetMetaResponse) error {
	return s.GetMeta(req, res)
}
