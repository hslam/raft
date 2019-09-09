package raft

import (
	"time"
	"sync"
	"errors"
)

type Log struct {
	mu 						sync.Mutex
	batchMu 				sync.Mutex
	node					*Node
	indexs 					*Index
	ret						uint64
	entries					[]*Entry
	entryChan             	chan *Entry
	readyEntries			[]*Entry
	maxBatch				int
	appendEntriesTicker		*time.Ticker
	compactionTicker		*time.Ticker
}

func newLog(node *Node) *Log {
	log:=&Log{
		node:					node,
		indexs:					newIndex(node),
		entryChan:				make(chan *Entry,DefaultMaxCacheEntries),
		readyEntries:			make([]*Entry,0),
		maxBatch:				DefaultMaxBatch,
		appendEntriesTicker:	time.NewTicker(DefaultMaxDelay),
		compactionTicker:	time.NewTicker(DefaultCompactionTick),
	}
	go log.run()
	return log
}
func (log *Log) checkIndex(index uint64)bool {
	if index==0{
		return false
	}
	length:=len(log.entries)
	if length==0{
		return false
	}
	if index<log.entries[0].Index||index>log.entries[length-1].Index{
		return false
	}
	return true
}
func (log *Log) lookup(index uint64)*Entry {
	log.mu.Lock()
	defer log.mu.Unlock()
	if !log.checkIndex(index){
		return nil
	}
	for i:=len(log.entries)-1;i>=0;i--{
		entry:= log.entries[i]
		if entry.Index==index{
			return entry
		}
	}
	return nil
}
func (log *Log) lookupLast(index uint64)*Entry {
	log.mu.Lock()
	defer log.mu.Unlock()
	if !log.checkIndex(index){
		return nil
	}
	length:=len(log.entries)
	if length<2{
		return nil
	}
	var cur int
	for i:=length-1;i>=0;i--{
		entry:= log.entries[i]
		if entry.Index==index{
			cur=i
			break
		}
	}
	if cur<1{
		return nil
	}
	return log.entries[cur-1]
}
func (log *Log) lookupNext(index uint64)*Entry {
	log.mu.Lock()
	defer log.mu.Unlock()
	length:=len(log.entries)
	if length<2{
		return nil
	}
	var cur int
	for i:=length-1;i>=0;i--{
		entry:= log.entries[i]
		if entry.Index==index{
			cur=i
			break
		}
	}
	if cur>length-2{
		return nil
	}
	return log.entries[cur+1]
}
func (log *Log) deleteAfter(index uint64) {
	log.mu.Lock()
	defer log.mu.Unlock()
	length:=len(log.entries)
	if length==0{
		return
	}
	if index<=1&&length>0{
		log.entries=log.entries[length:]
		return
	}
	for i:=length-1;i>=0;i--{
		entry:= log.entries[i]
		if entry.Index==index{
			log.entries=log.entries[:i]
			break
		}
	}
	log.save()
	Tracef("Log.deleteAfter %s delete %d and after",log.node.address, index)
}

func (log *Log) copyAfter(index uint64,max int)(entries []*Entry) {
	log.mu.Lock()
	defer log.mu.Unlock()
	length:=len(log.entries)
	if length==0{
		return
	}
	if index<=1&&length>0{
		if max<=length{
			entries=log.entries[0:max]
		}else {
			entries=log.entries[:]
		}
		return
	}
	for i:=len(log.entries)-1;i>=0;i--{
		entry:= log.entries[i]
		if entry.Index==index{
			if i+max<=len(log.entries){
				entries=log.entries[i:i+max]
			}else {
				entries=log.entries[i:]
			}
			break
		}
	}
	return
}
func (log *Log) copyRange(startIndex uint64,endIndex uint64) []*Entry{
	if startIndex>endIndex{
		return nil
	}
	length:=len(log.entries)
	if length==0{
		return nil
	}
	if endIndex<log.entries[0].Index||startIndex>log.entries[length-1].Index{
		return nil
	}
	var entries []*Entry
	for i:=0;i<len(log.entries);i++{
		entry:=log.entries[i]
		if entry.Index<startIndex{
			continue
		}else if entry.Index>endIndex{
			break
		}
		entries=append(entries, entry)
	}
	return entries
}
func (log *Log) applyCommited() {
	log.mu.Lock()
	defer log.mu.Unlock()
	length:=len(log.entries)
	if length==0{
		return
	}
	var startIndex =maxUint64(log.node.stateMachine.lastApplied,log.entries[0].Index)
	var endIndex =log.node.commitIndex
	log.applyCommitedRange(startIndex,endIndex)
}
func (log *Log) applyCommitedBefore(index uint64) {
	log.mu.Lock()
	defer log.mu.Unlock()
	if !log.checkIndex(index){
		return
	}
	var startIndex =maxUint64(log.node.stateMachine.lastApplied,log.entries[0].Index)
	var endIndex =index
	log.applyCommitedRange(startIndex,endIndex)
}
func (log *Log) applyCommitedRange(startIndex uint64,endIndex uint64) {
	if startIndex>endIndex{
		return
	}
	length:=len(log.entries)
	if length==0{
		return
	}
	if endIndex<log.entries[0].Index||startIndex>log.entries[length-1].Index{
		return
	}
	for i:=0;i<len(log.entries);i++{
		entry:=log.entries[i]
		if entry.Index<startIndex{
			continue
		}else if entry.Index>endIndex{
			break
		}
		command:=log.node.commandType.clone(entry.CommandType)
		err:=log.node.raftCodec.Decode(entry.Command,command)
		if err==nil{
			log.node.stateMachine.Apply(entry.Index,command)
		}
	}
}

func (log *Log) appendEntries(entries []*Entry) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.entries=append(log.entries,entries...)
	var metas []*Meta
	var data []byte
	data,metas,_=log.Encode(entries,log.ret)
	log.append(data)
	log.indexs.appendMetas(metas)
}

func (log *Log) append(b []byte) {
	log.node.storage.SeekWrite(DefaultLog,log.ret,b)
	log.ret+=uint64(len(b))
	if len(log.entries)>0{
		log.node.lastLogIndex=log.entries[len(log.entries)-1].Index
		log.node.lastLogTerm=log.entries[len(log.entries)-1].Term
	}
}
func (log *Log) save() {
	b,metas,_:=log.Encode(log.entries,0)
	log.indexs.ms.Metas=metas
	log.node.storage.SafeOverWrite(DefaultLog,b)
	log.indexs.save()
	log.ret=uint64(len(b))
	if len(log.entries)>0{
		log.node.lastLogIndex=log.entries[len(log.entries)-1].Index
		log.node.lastLogTerm=log.entries[len(log.entries)-1].Term
	}
}
func (log *Log) recover() error {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.indexs.recover()
	if !log.node.storage.Exists(DefaultLog){
		return nil
	}
	b, err := log.node.storage.Load(DefaultLog)
	if err != nil {
		return err
	}
	if log.ret,err = log.Decode(b); err != nil {
		return err
	}
	if len(log.entries)>0{
		log.node.lastLogIndex=log.entries[len(log.entries)-1].Index
		log.node.lastLogTerm=log.entries[len(log.entries)-1].Term
		log.node.recoverLogIndex=log.node.lastLogIndex
		log.node.nextIndex=log.node.lastLogIndex+1
	}
	return nil
}

func (log *Log)Decode(data []byte)(uint64,error)  {
	cnt:=len(log.indexs.ms.Metas)
	var log_ret uint64
	for j:=0;j<cnt;j++{
		ret:=log.indexs.ms.Metas[j].Ret
		offset:=log.indexs.ms.Metas[j].Offset
		b:=data[ret:ret+offset]
		entry:=&Entry{}
		err:=log.node.raftCodec.Decode(b,entry)
		if err!=nil{
			Errorf("Log.Decode %d %d %s",ret,offset,string(b))
		}
		log.entries=append(log.entries, entry)
		log_ret=uint64(ret+offset)
	}
	Tracef("Log.Decode %s entries length %d",log.node.address,cnt)
	return log_ret,nil
}

func (log *Log)Encode(entries []*Entry,ret uint64)([]byte,[]*Meta,error)  {
	var metas []*Meta
	var data []byte
	for i:=0;i<len(entries);i++{
		entry:=entries[i]
		b,_:=log.node.raftCodec.Encode(entry)
		data=append(data,b...)
		meta:=&Meta{}
		meta.Index=entry.Index
		meta.Term=entry.Term
		meta.Ret=ret
		length:=uint64(len(b))
		meta.Offset=length
		ret+=length
		metas=append(metas,meta)
	}
	return data,metas,nil
}
func (log *Log) compaction() error {
	log.mu.Lock()
	defer log.mu.Unlock()
	md5, err:=log.loadMd5()
	if err==nil&&md5!=""{
		if md5==log.node.storage.MD5(DefaultLog){
			return nil
		}
	}
	var compactionEntries =make(map[string]*Entry)
	var entries	=make([]*Entry,0)
	for i:=len(log.entries)-1;i>=0;i--{
		entry:= log.entries[i]
		key:=string(int32ToBytes(entry.CommandType))+entry.CommandId
		if _,ok:=compactionEntries[key];!ok{
			entries=append(entries,entry)
			compactionEntries[key]=entry
		}
	}
	for i,j:= 0,len(entries)-1;i<j;i,j=i+1,j-1{
		entries[i], entries[j] = entries[j], entries[i]
	}
	log.entries=entries

	b,metas,_:=log.Encode(log.entries,0)
	var lastLogSize ,_= log.node.storage.Size(DefaultLog)
	log.node.storage.SafeOverWrite(DefaultLog,b)
	log.indexs.ms.Metas=metas
	log.indexs.save()
	log.ret=uint64(len(b))
	if len(log.entries)>0{
		log.node.lastLogIndex=log.entries[len(log.entries)-1].Index
		log.node.lastLogTerm=log.entries[len(log.entries)-1].Term
	}
	var logSize ,_= log.node.storage.Size(DefaultLog)
	Tracef("Log.compaction %s LogSize %d==>%d",log.node.address,lastLogSize,logSize)
	log.saveMd5()
	Tracef("Log.compaction %s md5 %s==>%s",log.node.address,md5,log.node.storage.MD5(DefaultLog))
	return nil
}
func (log *Log) saveMd5() {
	md5:=log.node.storage.MD5(DefaultLog)
	if len(md5)>0{
		log.node.storage.SafeOverWrite(DefaultMd5,[]byte(md5))
	}
}

func (log *Log) loadMd5() (string,error) {
	if !log.node.storage.Exists(DefaultMd5){
		return "",errors.New("md5 file is not existed")
	}
	b, err := log.node.storage.Load(DefaultMd5)
	if err != nil {
		return "",err
	}
	return string(b),nil
}

func (log *Log) run()  {
	go func() {
		for entry := range log.entryChan {
			log.batchMu.Lock()
			log.readyEntries=append(log.readyEntries, entry)
			if len(log.readyEntries)>=log.maxBatch{
				entries:=log.readyEntries[:]
				log.readyEntries=nil
				log.readyEntries=make([]*Entry,0)
				log.ticker(entries)
			}
			log.batchMu.Unlock()
		}
	}()
	for{
		select {
		case <-log.appendEntriesTicker.C:
			log.batchMu.Lock()
			if len(log.readyEntries)>log.maxBatch{
				entries:=log.readyEntries[:log.maxBatch]
				log.readyEntries=log.readyEntries[log.maxBatch:]
				log.ticker(entries)
			}else  if len(log.readyEntries)>0{
				entries:=log.readyEntries[:]
				log.readyEntries=log.readyEntries[len(log.readyEntries):]
				log.ticker(entries)
			}
			log.batchMu.Unlock()
		case <-log.compactionTicker.C:
			//log.compaction()
		}
	}
}
func (log *Log) ticker(entries []*Entry) {
	log.appendEntries(entries)
}