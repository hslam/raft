package raft

func NewNoOperationCommand() Command {
	return &NoOperationCommand{}
}

func (c *NoOperationCommand) Type() int32 {
	return CommandTypeNoOperation
}

func (c *NoOperationCommand) UniqueID() string {
	return ""
}

func (c *NoOperationCommand) Do(context interface{}) (interface{}, error) {
	return true, nil
}
