package node

import (
	"hslam.com/mgit/Mort/raft"
)

func newSetCommand(key string, value string) raft.Command {
	c:=setCommandPool.Get().(*SetCommand)
	c.Key=key
	c.Value=value
	return c
}
func (c *SetCommand) Type()int32{
	return 1
}

func (c *SetCommand) UniqueID()string {
	return c.Key
}

func (c *SetCommand) Do(context interface{})(interface{},error){
	db := context.(*DB)
	db.Set(c.Key, c.Value)
	return nil,nil
}
