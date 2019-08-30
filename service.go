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

