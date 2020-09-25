// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

type service struct {
	node *Node
}

//RPCMethodName
func (s *service) RequestVote(req *RequestVoteRequest, res *RequestVoteResponse) error {
	return s.node.raft.HandleRequestVote(req, res)
}
func (s *service) AppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse) error {
	return s.node.raft.HandleAppendEntries(req, res)
}
func (s *service) InstallSnapshot(req *InstallSnapshotRequest, res *InstallSnapshotResponse) error {
	return s.node.raft.HandleInstallSnapshot(req, res)
}
func (s *service) QueryLeader(req *QueryLeaderRequest, res *QueryLeaderResponse) error {
	return s.node.proxy.HandleQueryLeader(req, res)
}
func (s *service) AddPeer(req *AddPeerRequest, res *AddPeerResponse) error {
	return s.node.proxy.HandleAddPeer(req, res)
}
func (s *service) RemovePeer(req *RemovePeerRequest, res *RemovePeerResponse) error {
	return s.node.proxy.HandleRemovePeer(req, res)
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
func (s *service) J(req *AddPeerRequest, res *AddPeerResponse) error {
	return s.AddPeer(req, res)
}
func (s *service) L(req *RemovePeerRequest, res *RemovePeerResponse) error {
	return s.RemovePeer(req, res)
}
