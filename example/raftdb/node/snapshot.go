package node

import (
	"encoding/json"
	"hslam.com/mgit/Mort/raft"
)

type Snapshot struct {}

func NewSnapshot() raft.Snapshot {
	return &Snapshot{}
}

func (s *Snapshot) Save(context interface{})([]byte,error){
	db := context.(*DB)
	var data map[string]string
	data=db.Data()
	return json.Marshal(data)
}

func (s *Snapshot) Recover(context interface{},snapshot []byte)(error){
	db := context.(*DB)
	var data map[string]string
	err:=json.Unmarshal(snapshot, &data)
	db.SetData(data)
	return err
}
