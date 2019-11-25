package main

import (
	"io"
	"encoding/json"
	"io/ioutil"
)
type Snapshot struct {}
func (s *Snapshot) Save(context interface{},w io.Writer) (int, error){
	ctx := context.(*Context)
	raw,err:=json.Marshal(ctx.data)
	if err!=nil{
		return 0,err
	}
	return w.Write(raw)
}

func (s *Snapshot) Recover(context interface{},r io.Reader) (int, error){
	ctx := context.(*Context)
	raw,err:=ioutil.ReadAll(r)
	err=json.Unmarshal(raw, ctx.data)
	return len(raw),err
}
