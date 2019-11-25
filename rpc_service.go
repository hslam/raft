package raft


type RPCService struct {
	node	*Node
}

//RPCMethodName
func (s *RPCService) RequestVote(req *RequestVoteRequest, res *RequestVoteResponse)error {
	return s.node.raft.HandleRequestVote(req,res)
}
func (s *RPCService) AppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse)error {
	return s.node.raft.HandleAppendEntries(req,res)
}
func (s *RPCService) InstallSnapshot(req *InstallSnapshotRequest,res *InstallSnapshotResponse)error {
	return s.node.raft.HandleInstallSnapshot(req,res)
}
func (s *RPCService) QueryLeader(req *QueryLeaderRequest,res *QueryLeaderResponse)error {
	return s.node.raft.HandleQueryLeader(req,res)
}
func (s *RPCService) AddPeer(req *AddPeerRequest,res *AddPeerResponse)error {
	return s.node.raft.HandleAddPeer(req,res)
}
func (s *RPCService) RemovePeer(req *RemovePeerRequest,res *RemovePeerResponse)error {
	return s.node.raft.HandleRemovePeer(req,res)
}
//RPCShortMethodName
func (s *RPCService) R(req *RequestVoteRequest, res *RequestVoteResponse)error {
	return s.RequestVote(req,res)
}
func (s *RPCService) A(req *AppendEntriesRequest, res *AppendEntriesResponse)error {
	return s.AppendEntries(req,res)
}
func (s *RPCService) I(req *InstallSnapshotRequest,res *InstallSnapshotResponse)error {
	return s.InstallSnapshot(req,res)
}
func (s *RPCService) Q(req *QueryLeaderRequest,res *QueryLeaderResponse)error {
	return s.QueryLeader(req,res)
}
func (s *RPCService) J(req *AddPeerRequest,res *AddPeerResponse)error {
	return s.AddPeer(req,res)
}
func (s *RPCService) L(req *RemovePeerRequest,res *RemovePeerResponse)error {
	return s.RemovePeer(req,res)
}

