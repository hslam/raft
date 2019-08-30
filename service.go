package raft


type Service struct {
	server	*Server
}

func (s *Service) R(req *RequestVoteRequest, res *RequestVoteResponse)error {
	return s.RequestVote(req,res)
}
func (s *Service) A(req *AppendEntriesRequest, res *AppendEntriesResponse)error {
	return s.AppendEntries(req,res)
}
func (s *Service) I(req *InstallSnapshotRequest,res *InstallSnapshotResponse)error {
	return s.InstallSnapshot(req,res)
}

func (s *Service) RequestVote(req *RequestVoteRequest, res *RequestVoteResponse)error {
	res.Term=s.server.currentTerm
	if req.Term<s.server.currentTerm{
		res.VoteGranted=false
		Tracef("Service.RequestVote %s vote %s VoteGranted %t",s.server.address,req.CandidateId,res.VoteGranted)
		return nil
	}
	s.server.resetLastRPCTime()
	if (s.server.votedFor==""||s.server.votedFor==req.CandidateId)&&s.server.commitIndex<=req.LastLogIndex&&s.server.currentTerm<=req.LastLogTerm{
		res.VoteGranted=true
		s.server.votedFor=req.CandidateId
		Tracef("Service.RequestVote %s vote %s VoteGranted %t",s.server.address,req.CandidateId,res.VoteGranted)
		return nil
	}
	return nil
}
func (s *Service) AppendEntries(req *AppendEntriesRequest, res *AppendEntriesResponse)error {
	res.Term=s.server.currentTerm
	if req.Term<s.server.currentTerm{
		res.Success=false
		return nil
	}
	if len(req.Entries)==0{
		if s.server.votedFor!=""&&s.server.votedFor!=req.LeaderId{
			res.Success=false
			return nil
		}
		res.Success=true
		s.server.votedFor=req.LeaderId
		s.server.leader=req.LeaderId
		s.server.currentTerm=req.Term
		s.server.resetLastRPCTime()
		//Tracef("Service.AppendEntries %s recv %s Heartbeat",s.server.address,req.LeaderId,)
		switch s.server.State() {
		case Follower:
		case Candidate:
			s.server.ChangeState(-1)
		case Leader:
			if req.Term>s.server.currentTerm{
				Tracef("Service.AppendEntries Leader %s recv %s Heartbeat",s.server.address,req.LeaderId,)
				s.server.ChangeState(-1)
			}

		}
		return nil
	}else {}
	var lastEntryIndex uint64
	for _,v :=range req.Entries {
		lastEntryIndex=v.Index
	}
	if req.LeaderCommit>s.server.commitIndex{
		s.server.commitIndex=min(req.LeaderCommit,lastEntryIndex)
		res.Success=true
		return nil
	}
	return nil
}
func (s *Service) InstallSnapshot(req *InstallSnapshotRequest,res *InstallSnapshotResponse)error {
	return nil
}

