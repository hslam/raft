package raft

//import (
//	"hslam.com/mgit/Mort/timer"
//	"sync"
//	"errors"
//)
//
//type Log struct {
//	mu 						sync.Mutex
//	batchMu 				sync.Mutex
//	node					*Node
//	indexs 					*Index
//	ret						uint64
//	entryChan             	chan *Entry
//	readyEntries			[]*Entry
//	maxBatch				int
//	appendEntriesTicker		*timer.Ticker
//	compactionTicker		*timer.Ticker
//}
//
//func newLog(node *Node) *Log {
//	log:=&Log{
//		node:					node,
//		indexs:					newIndex(node),
//		entryChan:				make(chan *Entry,DefaultMaxCacheEntries),
//		readyEntries:			make([]*Entry,0),
//		maxBatch:				DefaultMaxBatch,
//		appendEntriesTicker:	timer.NewFuncTicker(DefaultMaxDelay),
//		compactionTicker:	timer.NewTicker(DefaultCompactionTick),
//	}
//	go log.run()
//	return log
//}
//func (log *Log) checkIndex(index uint64)bool {
//	return log.indexs.checkIndex(index)
//}
//func (log *Log) lookup(index uint64)*Entry {
//	log.mu.Lock()
//	defer log.mu.Unlock()
//	meta:=log.indexs.lookup(index)
//	return log.read(meta)
//}
//func (log *Log) lookupLast(index uint64)*Entry {
//	log.mu.Lock()
//	defer log.mu.Unlock()
//	meta:=log.indexs.lookupLast(index)
//	return log.read(meta)
//}
//func (log *Log) lookupNext(index uint64)*Entry {
//	log.mu.Lock()
//	defer log.mu.Unlock()
//	meta:=log.indexs.lookupNext(index)
//	return log.read(meta)
//}
//func (log *Log) deleteAfter(index uint64) {
//	log.mu.Lock()
//	defer log.mu.Unlock()
//	log.indexs.deleteAfter(index)
//	length:=len(log.indexs.metas)
//	if length>0{
//		log.node.lastLogIndex=log.indexs.metas[length-1].Index
//		log.node.lastLogTerm=log.indexs.metas[length-1].Term
//		meta:=log.indexs.metas[len(log.indexs.metas)-1]
//		log.ret=meta.Ret+meta.Offset
//	}
//}
//
//func (log *Log) copyAfter(index uint64,max int)(entries []*Entry) {
//	log.mu.Lock()
//	defer log.mu.Unlock()
//	metas:=log.indexs.copyAfter(index,max)
//	return log.batchRead(metas)
//}
//func (log *Log) copyRange(startIndex uint64,endIndex uint64) []*Entry{
//	metas:=log.indexs.copyRange(startIndex,endIndex)
//	return log.batchRead(metas)
//}
//func (log *Log) applyCommited() {
//	log.mu.Lock()
//	defer log.mu.Unlock()
//	length:=len(log.indexs.metas)
//	if length==0{
//		return
//	}
//	var startIndex =maxUint64(log.node.stateMachine.lastApplied,log.indexs.metas[0].Index)
//	var endIndex =log.node.commitIndex
//
//	if endIndex-startIndex>DefaultMaxBatch{
//		index:=startIndex
//		for {
//			log.applyCommitedRange(index,index+DefaultMaxBatch)
//			index+=DefaultMaxBatch
//			if endIndex-index<=DefaultMaxBatch{
//				log.applyCommitedRange(index,endIndex)
//				break
//			}
//		}
//	}else {
//		log.applyCommitedRange(startIndex,endIndex)
//	}
//}
//
//func (log *Log) applyCommitedRange(startIndex uint64,endIndex uint64) {
//	entries:=log.copyRange(startIndex,endIndex)
//	for i:=0;i<len(entries);i++{
//		command:=log.node.commandType.clone(entries[i].CommandType)
//		err:=log.node.raftCodec.Decode(entries[i].Command,command)
//		if err==nil{
//			log.node.stateMachine.Apply(entries[i].Index,command)
//		}
//	}
//	//Tracef("Log.applyCommitedRange %s apply %d==>%d",log.node.address,startIndex,endIndex)
//}
//
//func (log *Log) appendEntries(entries []*Entry) {
//	log.mu.Lock()
//	defer log.mu.Unlock()
//	data,metas,_:=log.Encode(entries,log.ret)
//	log.append(data)
//	log.indexs.appendMetas(metas)
//	length:=len(log.indexs.metas)
//	if length>0{
//		log.node.lastLogIndex=log.indexs.metas[length-1].Index
//		log.node.lastLogTerm=log.indexs.metas[length-1].Term
//	}
//}
//func (log *Log) read(meta *Meta)*Entry {
//	if meta==nil{
//		return nil
//	}
//	b:=make([]byte,meta.Offset)
//	err:=log.node.storage.SeekRead(DefaultLog,meta.Ret,b)
//	if err!=nil{
//		return nil
//	}
//	entry:=&Entry{}
//	err=log.node.raftCodec.Decode(b,entry)
//	if err!=nil{
//		Errorf("Log.Decode %s",string(b))
//		return nil
//	}
//	return entry
//}
//
//func (log *Log) batchRead(metas []*Meta)[]*Entry {
//	if len(metas)==0{
//		return nil
//	}
//	cursor:=metas[0].Ret
//	offset:=metas[len(metas)-1].Ret+metas[len(metas)-1].Offset
//	b:=make([]byte,offset-cursor)
//	err:=log.node.storage.SeekRead(DefaultLog,cursor,b)
//	if err!=nil{
//		return nil
//	}
//	if err != nil {
//		return nil
//	}
//	var entries []*Entry
//	if entries,_,err = log.Decode(b,metas,cursor); err != nil {
//		return nil
//	}
//	return entries
//}
//
//func (log *Log) append(b []byte) {
//	log.node.storage.SeekWrite(DefaultLog,log.ret,b)
//	log.ret+=uint64(len(b))
//}
//
//func (log *Log) recover() error {
//	log.mu.Lock()
//	defer log.mu.Unlock()
//	log.indexs.recover()
//	if !log.node.storage.Exists(DefaultLog){
//		return nil
//	}
//	length:=len(log.indexs.metas)
//	if length>0{
//		log.node.lastLogIndex=log.indexs.metas[length-1].Index
//		log.node.lastLogTerm=log.indexs.metas[length-1].Term
//		log.node.recoverLogIndex=log.node.lastLogIndex
//		log.node.nextIndex=log.node.lastLogIndex+1
//		meta:=log.indexs.metas[len(log.indexs.metas)-1]
//		log.ret=meta.Ret+meta.Offset
//	}
//	return nil
//}
//
//func (log *Log)Decode(data []byte,metas []*Meta,cursor uint64,)([]*Entry,uint64,error)  {
//	length:=len(metas)
//	entries:=make([]*Entry,0,length)
//	var log_ret uint64
//	for j:=0;j<length;j++{
//		ret:=metas[j].Ret-cursor
//		offset:=metas[j].Offset
//		b:=data[ret:ret+offset]
//		entry:=&Entry{}
//		err:=log.node.raftCodec.Decode(b,entry)
//		if err!=nil{
//			Errorf("Log.Decode %d %d %s",ret,offset,string(b))
//		}
//		entries=append(entries, entry)
//		log_ret=uint64(ret+offset)
//	}
//	return entries,cursor+log_ret,nil
//}
//
//func (log *Log)Encode(entries []*Entry,ret uint64)([]byte,[]*Meta,error)  {
//	var metas []*Meta
//	var data []byte
//	for i:=0;i<len(entries);i++{
//		entry:=entries[i]
//		b,_:=log.node.raftCodec.Encode(entry)
//		data=append(data,b...)
//		meta:=&Meta{}
//		meta.Index=entry.Index
//		meta.Term=entry.Term
//		meta.Ret=ret
//		length:=uint64(len(b))
//		meta.Offset=length
//		ret+=length
//		metas=append(metas,meta)
//	}
//	return data,metas,nil
//}
//func (log *Log) compaction() error {
//	//log.mu.Lock()
//	//defer log.mu.Unlock()
//	//md5, err:=log.loadMd5()
//	//if err==nil&&md5!=""{
//	//	if md5==log.node.storage.MD5(DefaultLog){
//	//		return nil
//	//	}
//	//}
//	//var compactionEntries =make(map[string]*Entry)
//	//var entries	=make([]*Entry,0)
//	//for i:=len(log.entries)-1;i>=0;i--{
//	//	entry:= log.entries[i]
//	//	key:=string(int32ToBytes(entry.CommandType))+entry.CommandId
//	//	if _,ok:=compactionEntries[key];!ok{
//	//		entries=append(entries,entry)
//	//		compactionEntries[key]=entry
//	//	}
//	//}
//	//for i,j:= 0,len(entries)-1;i<j;i,j=i+1,j-1{
//	//	entries[i], entries[j] = entries[j], entries[i]
//	//}
//	//log.entries=entries
//	//
//	//b,metas,_:=log.Encode(log.entries,0)
//	//var lastLogSize ,_= log.node.storage.Size(DefaultLog)
//	//log.node.storage.SafeOverWrite(DefaultLog,b)
//	//log.indexs.metas=metas
//	//log.indexs.save()
//	//log.ret=uint64(len(b))
//	//if len(log.entries)>0{
//	//	log.node.lastLogIndex=log.entries[len(log.entries)-1].Index
//	//	log.node.lastLogTerm=log.entries[len(log.entries)-1].Term
//	//}
//	//var logSize ,_= log.node.storage.Size(DefaultLog)
//	//Tracef("Log.compaction %s LogSize %d==>%d",log.node.address,lastLogSize,logSize)
//	//log.saveMd5()
//	//Tracef("Log.compaction %s md5 %s==>%s",log.node.address,md5,log.node.storage.MD5(DefaultLog))
//	return nil
//}
//func (log *Log) saveMd5() {
//	md5:=log.node.storage.MD5(DefaultLog)
//	if len(md5)>0{
//		log.node.storage.SafeOverWrite(DefaultMd5,[]byte(md5))
//	}
//}
//
//func (log *Log) loadMd5() (string,error) {
//	if !log.node.storage.Exists(DefaultMd5){
//		return "",errors.New("md5 file is not existed")
//	}
//	b, err := log.node.storage.Load(DefaultMd5)
//	if err != nil {
//		return "",err
//	}
//	return string(b),nil
//}
//
//func (log *Log) run()  {
//	go func() {
//		for entry := range log.entryChan {
//			log.batchMu.Lock()
//			log.readyEntries=append(log.readyEntries, entry)
//			if len(log.readyEntries)>=log.maxBatch{
//				entries:=log.readyEntries[:]
//				log.readyEntries=nil
//				log.readyEntries=make([]*Entry,0)
//				log.ticker(entries)
//			}
//			log.batchMu.Unlock()
//		}
//	}()
//	log.appendEntriesTicker.Tick(func() {
//		log.batchMu.Lock()
//		if len(log.readyEntries)>log.maxBatch{
//			entries:=log.readyEntries[:log.maxBatch]
//			log.readyEntries=log.readyEntries[log.maxBatch:]
//			log.ticker(entries)
//		}else  if len(log.readyEntries)>0{
//			entries:=log.readyEntries[:]
//			log.readyEntries=log.readyEntries[len(log.readyEntries):]
//			log.ticker(entries)
//		}
//		log.batchMu.Unlock()
//	})
//	for{
//		select {
//		case <-log.compactionTicker.C:
//			//log.compaction()
//		}
//	}
//}
//func (log *Log) ticker(entries []*Entry) {
//	log.appendEntries(entries)
//}