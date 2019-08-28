package raft

type Service struct {}

func (s *Service) R(req *RequestVoteRequest, res *RequestVoteResponse)error {
	return nil
}
func (s *Service) A(req *AppendEntriesRequest, res *AppendEntriesResponse)error {
	res.Term=req.Term
	return nil
}
func (s *Service) I(req *InstallSnapshotRequest,res *InstallSnapshotResponse)error {
	return nil
}

