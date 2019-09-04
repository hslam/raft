package raft


type Snapshot struct {
	node 	*Node
}

func newSnapshot(node *Node)*Snapshot {
	s:=&Snapshot{
		node:node,
	}
	return s
}
func (s *Snapshot)Compact() {

}