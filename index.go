package raft

import (
	"sync"
)
type Index struct {
	mu 				sync.RWMutex
	node			*Node
	ms				*MetaStorage
	ret				uint64
}
func newIndex(node *Node) *Index {
	i:=&Index{
		node:node,
		ms:&MetaStorage{Metas:make([]*Meta,0)},
	}
	return i
}
func (i *Index) appendMetas(metas []*Meta) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.ms.Metas=append(i.ms.Metas,metas...)
	data,_:=i.Encode(metas)
	i.append(data)
}

func (i *Index) append(b []byte) {
	i.node.storage.SeekWrite(DefaultIndex,i.ret,b)
	i.ret+=uint64(len(b))
}
func (i *Index) save() {
	b,_:=i.Encode(i.ms.Metas)
	i.node.storage.SafeOverWrite(DefaultIndex,b)
	i.ret=uint64(len(b))
}
func (i *Index) recover() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if !i.node.storage.Exists(DefaultIndex){
		return nil
	}
	b, err := i.node.storage.Load(DefaultIndex)
	if err != nil {
		return err
	}
	if err = i.Decode(b); err != nil {
		return err
	}
	i.ret=uint64(len(i.ms.Metas)*32)
	return nil
}
func (i *Index)Decode(data []byte)error  {
	cnt:=len(data)/32
	for j:=0;j<cnt;j++{
		b:=data[j*32:j*32+32]
		meta:=&Meta{}
		meta.Decode(b)
		i.ms.Metas=append(i.ms.Metas, meta)
	}
	Tracef("Index.Decode %s Metas length %d",i.node.address,cnt)
	return nil
}
func (i *Index)Encode(metas []*Meta)([]byte,error)  {
	var data=[]byte{}
	for j:=0;j<len(metas);j++{
		b:=metas[j].Encode()
		data=append(data,b...)
	}
	return data,nil
}
func (i *Index)Lookup()  {

}