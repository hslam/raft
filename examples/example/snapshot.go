package main

import (
	"io"
	"io/ioutil"
)
type Snapshot struct {}
func (s *Snapshot) Save(context interface{},w io.Writer) (int, error){
	ctx := context.(*Context)
	return w.Write([]byte(ctx.Get()))
}

func (s *Snapshot) Recover(context interface{},r io.Reader) (int, error){
	ctx := context.(*Context)
	raw,err:=ioutil.ReadAll(r)
	if err!=nil{
		return 0,err
	}
	ctx.Set(string(raw))
	return len(raw),err
}
