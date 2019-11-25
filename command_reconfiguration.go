package raft


func NewReconfigurationCommand() Command {
	return &ReconfigurationCommand{}
}
func (c *ReconfigurationCommand) Type() int32 {
	return CommandTypeReconfiguration
}

func (c *ReconfigurationCommand) UniqueID() string{
	return ""
}

func (c *ReconfigurationCommand) Do(context interface{})(interface{},error){
	node := context.(*Node)
	node.stateMachine.configuration.reconfiguration()
	return nil, nil
}
