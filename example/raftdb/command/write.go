package command

import (
	"hslam.com/mgit/Mort/raft"
	"hslam.com/mgit/Mort/raft/example/raftdb/db"
)

type WriteCommand struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewWriteCommand(key string, value string) *WriteCommand {
	return &WriteCommand{
		Key:   key,
		Value: value,
	}
}
func (c *WriteCommand) Type()uint32{
	return 1
}

func (c *WriteCommand) UniqueID()string {
	return c.Key
}

func (c *WriteCommand) Do(node *raft.Node)(interface{},error){
	db := node.Context().(*db.DB)
	db.Put(c.Key, c.Value)
	return nil,nil
}
