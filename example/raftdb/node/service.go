package node

type Service struct {
	node *Node
}

func (s *Service)Set(req *Request, res *Response) error {
	_, err := s.node.raft_node.Do(newSetCommand(req.Key, req.Value))
	if err==nil{
		res.Ok=true
	}
	return nil
}
