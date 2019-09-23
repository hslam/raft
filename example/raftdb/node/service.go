package node

import "hslam.com/mgit/Mort/raft"

type Service struct {
	node *Node
}

func (s *Service)Set(req *Request, res *Response) error {
	_, err := s.node.raft_node.Do(newSetCommand(req.Key, req.Value))
	if err==nil{
		res.Ok=true
		return nil
	}
	if err==raft.ErrNotLeader{
		leader:=s.node.raft_node.Leader()
		if leader!=""{
			res.Leader=leader
		}
	}
	return err
}
