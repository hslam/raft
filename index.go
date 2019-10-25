package raft

import (
	"sync"
	"errors"
	"math"
)

const (
	metaSize  = 32
)
var (
	metaBytesPool			*sync.Pool
)
func init() {
	metaBytesPool= &sync.Pool{
		New: func() interface{} {
			return make([]byte,metaSize)
		},
	}
}
type Index struct {
	node			*Node
	metaPool 		*sync.Pool
	firstIndex 		uint64
}
func newIndex(node *Node) *Index {
	i:=&Index{
		node:node,
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
	meta.Position=0
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
func (i *Index)check(metas []*Meta)(bool)  {
	lastIndex:=metas[0].Index
	for i:=1;i<len(metas);i++{
		if metas[i].Index==lastIndex+1{
			lastIndex=metas[i].Index
		}else {
			return false
		}
	}
	return true
}
func (i *Index) lookup(index uint64)*Meta {
	if !i.checkIndex(index){
		return nil
	}
	return i.readIndex(index)
}
func (i *Index) lookupLast(index uint64)*Meta {
	lastIndex:=index-1
	if !i.checkIndex(lastIndex){
		return nil
	}
	return i.readIndex(lastIndex)
}
func (i *Index) lookupNext(index uint64)*Meta {
	nextIndex:=index+1
	if !i.checkIndex(nextIndex){
		return nil
	}
	return i.readIndex(nextIndex)
}
func (i *Index) deleteAfter(index uint64){
	i.node.storage.Truncate(DefaultIndex,(index-1)*metaSize)
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
	return i.batchReadIndex(startIndex,endIndex)
}
func (i *Index) readIndex(index uint64)*Meta {
	if index<1{
		return nil
	}
	position:=(index-1)*32
	return i.read(position)
}
func (i *Index) read(position uint64)*Meta {
	b:=metaBytesPool.Get().([]byte)
	_,err:=i.node.storage.SeekRead(DefaultIndex,position,b)
	if err!=nil{
		return nil
	}
	meta:=i.getEmtyMeta()
	meta.Decode(b)
	metaBytesPool.Put(b)
	return meta
}
func (i *Index) batchReadIndex(startIndex uint64,endIndex uint64)[]*Meta {
	if startIndex<1||endIndex<1||startIndex>endIndex{
		return nil
	}
	length:=i.length()
	if length<1||startIndex>length||endIndex>length{
		return nil
	}
	position:=(startIndex-1)*metaSize
	offset:=endIndex*metaSize
	return i.batchRead(position,offset)
}
func (i *Index) batchRead(position uint64,offset uint64)[]*Meta {
	b:=make([]byte,offset-position)
	if _,err:=i.node.storage.SeekRead(DefaultIndex,position,b);err!=nil{
		return nil
	}
	if metas,err := i.Decode(b); err == nil {
		return metas
	}
	return nil
}

func (i *Index) append(b []byte) {
	lastLogIndex:=i.node.lastLogIndex
	var ret uint64=0
	if lastLogIndex>0{
		ret=uint64(lastLogIndex*metaSize)
	}else {
		ret=0
	}
	i.node.storage.SeekWrite(DefaultIndex,ret,b)
}

func (i *Index) load() error {
	if !i.node.storage.Exists(DefaultIndex){
		return errors.New(DefaultIndex+" file is not existed")
	}
	size,err:=i.node.storage.Size(DefaultIndex)
	if err!=nil{
		return err
	}
	var position uint64=0
	meta:=i.read(position)
	i.node.firstLogIndex=meta.Index
	i.putEmtyMeta(meta)
	position=uint64(size)-32
	meta=i.read(position)
	i.node.lastLogIndex=meta.Index
	i.putEmtyMeta(meta)
	return nil
}


func (i *Index) length() uint64 {
	return i.node.lastLogIndex
}

func (i *Index)Decode(data []byte)([]*Meta,error)  {
	length:=len(data)/metaSize
	metas:=make([]*Meta,0,length)
	for j:=0;j<length;j++{
		b:=data[j*metaSize:j*metaSize+metaSize]
		meta:=i.getEmtyMeta()
		meta.Decode(b)
		metas=append(metas, meta)
	}
	//Tracef("Index.Decode %s Metas length %d",i.node.address,length)
	return metas,nil
}
func (i *Index)Encode(metas []*Meta)([]byte,error)  {
	var data=make([]byte,0,len(metas)*metaSize)
	for j:=0;j<len(metas);j++{
		b:=metas[j].Encode()
		data=append(data,b...)
	}
	return data,nil
}
func (i *Index)Segment(index uint64)(uint64)  {
	return uint64(math.Ceil(float64(index)/DefaultMaxEntriesPerFile)*DefaultMaxEntriesPerFile)
}
