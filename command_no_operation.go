package raft

type noOperationCommand struct {
}

func newNoOperationCommand() Command {
	return &noOperationCommand{}
}

func (c *noOperationCommand) Type()int32{
	return CommandTypeNoOperation
}

func (c *noOperationCommand) UniqueID()string {
	return ""
}

func (c *noOperationCommand) Do(context interface{})(interface{},error){
	return true,nil
}
