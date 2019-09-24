package raft

import (
	"sync"
)
var (
	metaBytesPool			*sync.Pool
)
func init() {
	metaBytesPool= &sync.Pool{
		New: func() interface{} {
			return make([]byte,32)
		},
	}
}
type Index struct {
	node			*Node
	//metas			[]*Meta
	ret				uint64
	metaPool 		*sync.Pool
}
func newIndex(node *Node) *Index {
	i:=&Index{
		node:node,
		//metas:make([]*Meta,0),
	}
	i.metaPool= &sync.Pool{
		New: func() interface{} {
			return &Meta{}
		},
	}
	return i
}
func (i *Index) getEmtyMeta()*Meta {
	return i.metaPool.Get().(*Meta)
}
func (i *Index) putEmtyMeta(meta *Meta) {
	meta.Index=0
	meta.Term=0
	meta.Ret=0
	meta.Offset=0
	i.metaPool.Put(meta)
}
func (i *Index) putEmtyMetas(metas []*Meta) {
	for _,meta:=range metas{
		i.putEmtyMeta(meta)
	}
}
func (i *Index) appendMetas(metas []*Meta) {
	data,_:=i.Encode(metas)
	i.append(data)
}
func (i *Index) checkIndex(index uint64)bool {
	if index==0{
		return false
	}
	length:=i.length()
	if length==0{
		return false
	}
	if index<1||index>length{
		return false
	}
	return true
}
func (i *Index) lookup(index uint64)*Meta {
	if !i.checkIndex(index){
		return nil
	}
	return i.read(index)
}
func (i *Index) lookupLast(index uint64)*Meta {
	lastIndex:=index-1
	if !i.checkIndex(lastIndex){
		return nil
	}
	return i.read(lastIndex)
}
func (i *Index) lookupNext(index uint64)*Meta {
	nextIndex:=index+1
	if !i.checkIndex(nextIndex){
		return nil
	}
	return i.read(nextIndex)
}
func (i *Index) deleteAfter(index uint64){
	if index<=1{
		i.node.lastLogIndex.Set(0)
		return
	}
	i.node.lastLogIndex.Set(index)
	Tracef("Index.deleteAfter %s delete %d and after",i.node.address, index)
}
func (i *Index) copyAfter(index uint64,max int)(metas []*Meta) {
	length:=i.length()
	if length==0{
		return
	}
	var startIndex uint64
	var endIndex uint64
	if index<1{
		startIndex=1
	}else {
		startIndex=index
	}
	if length<startIndex+uint64(max){
		endIndex=length
	}else {
		endIndex=startIndex+uint64(max)
	}
	return i.copyRange(startIndex,endIndex)
}
func (i *Index) copyRange(startIndex uint64,endIndex uint64) []*Meta{
	return i.batchRead(startIndex,endIndex)
}
func (i *Index) read(index uint64)*Meta {
	if index<1{
		return nil
	}
	ret:=(index-1)*32
	b:=metaBytesPool.Get().([]byte)
	err:=i.node.storage.SeekRead(DefaultIndex,ret,b)
	if err!=nil{
		return nil
	}
	meta:=i.getEmtyMeta()
	meta.Decode(b)
	metaBytesPool.Put(b)
	return meta
}
func (i *Index) batchRead(startIndex uint64,endIndex uint64)[]*Meta {
	if startIndex<1||endIndex<1||startIndex>endIndex{
		return nil
	}
	length:=i.length()
	if length<1||startIndex>length||endIndex>length{
		return nil
	}
	cursor:=(startIndex-1)*32
	offset:=endIndex*32
	b:=make([]byte,offset-cursor)
	if err:=i.node.storage.SeekRead(DefaultIndex,cursor,b);err!=nil{
		return nil
	}
	if metas,err := i.Decode(b); err == nil {
		return metas
	}
	return nil
}

func (i *Index) append(b []byte) {
	i.node.storage.SeekWrite(DefaultIndex,i.ret,b)
	i.ret+=uint64(len(b))
}

func (i *Index) recover() error {
	if !i.node.storage.Exists(DefaultIndex){
		return nil
	}
	i.node.lastLogIndex.load()
	lastLogIndex:=i.node.lastLogIndex.Id()
	if lastLogIndex>0{
		i.ret=uint64(lastLogIndex*32)
	}
	return nil
}

func (i *Index) length() uint64 {
	return i.node.lastLogIndex.Id()
}

func (i *Index)Decode(data []byte)([]*Meta,error)  {
	length:=len(data)/32
	metas:=make([]*Meta,0,length)
	for j:=0;j<length;j++{
		b:=data[j*32:j*32+32]
		meta:=i.getEmtyMeta()
		meta.Decode(b)
		metas=append(metas, meta)
	}
	//Tracef("Index.Decode %s Metas length %d",i.node.address,length)
	return metas,nil
}
func (i *Index)Encode(metas []*Meta)([]byte,error)  {
	var data=make([]byte,0,len(metas)*32)
	for j:=0;j<len(metas);j++{
		b:=metas[j].Encode()
		data=append(data,b...)
	}
	return data,nil
}
