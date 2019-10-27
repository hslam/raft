package raft

import (
	"sync"
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
	name 			string
	tmpName 		string
	flushName 		string
}
func newIndex(node *Node) *Index {
	i:=&Index{
		node:node,
		name:DefaultIndex,
		tmpName:DefaultIndex+DefaultTmp,
		flushName:DefaultIndex+DefaultFlush,
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
	if i.length()==0{
		return false
	}
	if index<i.node.firstLogIndex||index>i.node.lastLogIndex{
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
	return i.lookup(index-1)
}
func (i *Index) lookupNext(index uint64)*Meta {
	return i.lookup(index+1)
}
func (i *Index) deleteAfter(index uint64){
	if !i.checkIndex(index){
		return
	}
	i.node.storage.Truncate(i.name,(index-1)*metaSize)
}
func (i *Index) deleteBefore(index uint64)uint64{
	if !i.checkIndex(index){
		return 0
	}
	cutoff:=i.flush(index)
	if i.firstIndexName(i.flushName)==index+1||i.lastIndexName(i.flushName)==i.node.lastLogIndex{
		i.node.storage.Rename(i.name,i.tmpName)
		i.node.storage.Rename(i.flushName,i.name)
		i.load()
		return cutoff
	}else {
		if i.node.storage.Exists(i.flushName){
			i.node.storage.Rm(i.flushName)
		}
		return 0
	}
	return cutoff
}
func (i *Index) flush(index uint64)uint64{
	meta:=i.lookupNext(index)
	cutoff:=meta.Position
	firstLogIndex:=meta.Index
	var startIndex =meta.Index
	var endIndex =i.node.lastLogIndex
	if startIndex>endIndex{
		return 0
	}
	max:=uint64(DefaultReadFileBufferSize/metaSize)
	if endIndex-startIndex>max{
		batch_index:=startIndex
		for {
			i.batchFlush(batch_index,batch_index+max,firstLogIndex,cutoff)
			batch_index+=max
			if endIndex-batch_index<=max{
				i.batchFlush(batch_index,endIndex,firstLogIndex,cutoff)
				break
			}
		}
	}else {
		i.batchFlush(startIndex,endIndex,firstLogIndex,cutoff)
	}
	return cutoff
}
func (i *Index) batchFlush(startIndex uint64,endIndex uint64,firstLogIndex,cutoff uint64)error{
	position:=i.position(startIndex)
	offset:=(endIndex-startIndex+1)*metaSize
	src:=i.batchReadBytes(position,offset)
	metas,err:=i.Decode(src)
	if err!=nil{
		return err
	}
	for j:=0;j<len(metas);j++{
		metas[j].Position-=cutoff
	}
	src,err=i.Encode(metas)
	if err!=nil{
		return err
	}
	var ret uint64=0
	if startIndex>0{
		ret=uint64((startIndex-firstLogIndex)*metaSize)
	}else {
		ret=0
	}
	i.node.storage.SeekWrite(i.flushName,ret,src)
	return nil
}
func (i *Index) startIndex(index uint64)uint64{
	return maxUint64(i.node.firstLogIndex,index)
}
func (i *Index) endIndex(index uint64)uint64{
	return minUint64(index,i.node.lastLogIndex)
}
func (i *Index) copyAfter(index uint64,max int)(metas []*Meta) {
	if i.length()==0{
		return
	}
	startIndex:=i.startIndex(index)
	endIndex:=i.endIndex(startIndex+uint64(max))
	return i.copyRange(startIndex,endIndex)
}

func (i *Index) copyRange(startIndex uint64,endIndex uint64) []*Meta{
	return i.batchReadIndex(startIndex,endIndex)
}
func (i *Index) position(index uint64)uint64{
	return (index-i.node.firstLogIndex)*metaSize
}

func (i *Index) readIndex(index uint64)*Meta {
	if !i.checkIndex(index){
		return nil
	}
	position:=i.position(index)
	return i.read(position)
}
func (i *Index) read(position uint64)*Meta {
	return i.readName(i.name,position)
}
func (i *Index) readName(name string,position uint64)*Meta {
	b:=metaBytesPool.Get().([]byte)
	_,err:=i.node.storage.SeekRead(name,position,b)
	if err!=nil{
		return nil
	}
	meta:=i.getEmtyMeta()
	meta.Decode(b)
	metaBytesPool.Put(b)
	return meta
}
func (i *Index) batchReadIndex(startIndex uint64,endIndex uint64)[]*Meta {
	if !i.checkIndex(startIndex)||!i.checkIndex(endIndex)||startIndex>endIndex||i.length()==0{
		return nil
	}
	position:=i.position(startIndex)
	offset:=(endIndex-startIndex+1)*metaSize
	return i.batchRead(position,offset)
}
func (i *Index) batchRead(position uint64,offset uint64)[]*Meta {
	b:=i.batchReadBytes(position,offset)
	if b==nil{
		return nil
	}
	if metas,err := i.Decode(b); err == nil {
		return metas
	}
	return nil
}
func (i *Index) batchReadBytes(position uint64,offset uint64)[]byte{
	b:=make([]byte,offset)
	if _,err:=i.node.storage.SeekRead(i.name,position,b);err!=nil{
		return nil
	}
	return b
}

func (i *Index) append(b []byte) {
	var ret uint64=0
	if i.node.lastLogIndex>0{
		ret=uint64((i.node.lastLogIndex-i.node.firstLogIndex+1)*metaSize)
	}else {
		ret=0
	}
	i.node.storage.SeekWrite(i.name,ret,b)
}

func (i *Index) load() error {
	i.node.firstLogIndex=i.firstIndex()
	i.node.lastLogIndex=i.lastIndex()
	return nil
}
func (i *Index) firstIndex() (uint64) {
	return i.firstIndexName(i.name)
}
func (i *Index) firstIndexName(name string) (uint64) {
	if !i.node.storage.Exists(name){
		return 1
	}
	var firstIndex uint64
	meta:=i.readName(name,0)
	firstIndex=meta.Index
	i.putEmtyMeta(meta)
	if firstIndex==0{
		firstIndex=1
	}
	return firstIndex
}
func (i *Index) lastIndex() (uint64) {
	return i.lastIndexName(i.name)
}
func (i *Index) lastIndexName(name string) (uint64) {
	if !i.node.storage.Exists(name){
		return 0
	}
	size,err:=i.node.storage.Size(name)
	if err!=nil&&size<metaSize{
		return 0
	}
	var lastIndex uint64
	meta:=i.readName(name,uint64(size)-metaSize)
	lastIndex=meta.Index
	i.putEmtyMeta(meta)
	return lastIndex
}
func (i *Index) length() uint64 {
	return i.node.lastLogIndex-i.node.firstLogIndex+1
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
