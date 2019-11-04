package node

import (
	"io"
	"io/ioutil"
	"encoding/json"
	"hslam.com/git/x/raft"
)

type Snapshot struct {}

func NewSnapshot() raft.Snapshot {
	return &Snapshot{}
}

func (s *Snapshot) Save(context interface{},w io.Writer) (int, error){
	db := context.(*DB)
	var data map[string]string
	data=db.Data()
	raw,err:=json.Marshal(data)
	if err!=nil{
		return 0,err
	}
	return w.Write(raw)
}

func (s *Snapshot) Recover(context interface{},r io.Reader) (int, error){
	db := context.(*DB)
	var data map[string]string
	raw,err:=ioutil.ReadAll(r)
	err=json.Unmarshal(raw, &data)
	if err!=nil{
		return len(raw),err
	}
	db.SetData(data)
	return len(raw),nil
}
