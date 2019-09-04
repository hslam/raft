package raft

type StateMachine struct {
	node 							*Node
	lastApplied						uint64
	cmds 							*Commands

}
func newStateMachine(node *Node)*StateMachine {
	s:=&StateMachine{
		node:node,
		cmds:newCommands(),
	}
	return s
}
func (s *StateMachine)Apply(command Command) {
	s.cmds.AddCommand(command)
	command.Do(s.node)
}