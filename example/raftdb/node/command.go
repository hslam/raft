package node

import (
	"hslam.com/mgit/Mort/raft"
)

//type SetCommand struct {
//	Key   	string		`json:"k"`
//	Value 	string		`json:"v"`
//}

func newSetCommand(key string, value string) raft.Command {
	return &SetCommand{
		Key:   key,
		Value: value,
	}
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
