package raft

import (
	"errors"
	"sync"
	"time"
)

type Log struct {
	mu 						sync.Mutex
	batchMu 				sync.Mutex
	pauseMu 				sync.Mutex
	node					*Node
	indexs 					*Index
	size					uint64
	entryChan             	chan *Entry
	readyEntries			[]*Entry
	maxBatch				int
	compactionTicker		*time.Ticker
	entryPool 				*sync.Pool
	stop					chan bool
	finish 					chan bool
	work 					bool
	paused 					bool
	name 			string
	tmpName 		string
	flushName 		string

}

func newLog(node *Node) *Log {
	log:=&Log{
		node:					node,
		indexs:					newIndex(node),
		entryChan:				make(chan *Entry,DefaultMaxCacheEntries),
		readyEntries:			make([]*Entry,0),
		maxBatch:				DefaultMaxBatch,
		compactionTicker:		time.NewTicker(DefaultCompactionTick),
		stop:make(chan bool,1),
		finish:make(chan bool,1),
		work:true,
		name:DefaultLog,
		tmpName:DefaultLog+DefaultTmp,
		flushName:DefaultLog+DefaultFlush,
	}
	log.entryPool= &sync.Pool{
		New: func() interface{} {
			return &Entry{}
		},
	}
	go log.run()
	return log
}
func (log *Log) getEmtyEntry()*Entry {
	return log.entryPool.Get().(*Entry)
}
func (log *Log) putEmtyEntry(entry *Entry) {
	entry.Index=0
	entry.Term=0
	entry.Command=[]byte{}
	entry.CommandType=0
	entry.CommandId=""
	log.entryPool.Put(entry)
}
func (log *Log) putEmtyEntries(entries []*Entry) {
	for _,entry:=range entries{
		log.putEmtyEntry(entry)
	}
}
func (log *Log) pause(p bool){
	log.pauseMu.Lock()
	defer log.pauseMu.Unlock()
	log.paused=p
}
func (log *Log) isPaused()bool{
	log.pauseMu.Lock()
	defer log.pauseMu.Unlock()
	return log.paused
}
func (log *Log) checkPaused(){
	for{
		if !log.isPaused(){
			break
		}
		time.Sleep(time.Millisecond*100)
	}
}
func (log *Log) checkIndex(index uint64)bool {
	return log.indexs.checkIndex(index)
}

func (log *Log) lookup(index uint64)*Entry {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	meta:=log.indexs.lookup(index)
	return log.read(meta)
}

func (log *Log) consistencyCheck(index uint64,term uint64)(ok bool) {
	if index==0{
		return false
	}
	if index>log.node.lastLogIndex{
		return false
	}
	if index==log.node.lastLogIndex{
		if term==log.node.lastLogTerm{
			return true
		}
	}
	entry:=log.node.log.lookup(index)
	if  entry==nil{
		return false
	}
	if entry.Term!=term{
		log.node.log.deleteAfter(index)
		log.node.nextIndex=log.node.lastLogIndex+1
		return false
	}
	return true
}
func (log *Log)check(entries []*Entry)(bool)  {
	lastIndex:=entries[0].Index
	for i:=1;i<len(entries);i++{
		if entries[i].Index==lastIndex+1{
			lastIndex=entries[i].Index
		}else {
			return false
		}
	}
	return true
}
func (log *Log) lookupLast(index uint64)*Entry {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	meta:=log.indexs.lookupLast(index)
	return log.read(meta)
}
func (log *Log) lookupNext(index uint64)*Entry {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	meta:=log.indexs.lookupNext(index)
	return log.read(meta)
}
func (log *Log) deleteAfter(index uint64) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.pause(true)
	defer log.pause(false)
	if index==0{
		log.node.lastLogIndex=0
		return
	}
	log.node.lastLogIndex=index-1
	log.indexs.deleteAfter(index)
	lastLogIndex:=log.node.lastLogIndex
	if lastLogIndex>0{
		meta:=log.indexs.lookup(lastLogIndex)
		log.node.lastLogTerm=meta.Term
		log.size=meta.Position+meta.Offset
	}else {
		log.node.lastLogTerm=0
		log.size=0
	}
	log.node.storage.Truncate(DefaultLog,log.size)
}
func (log *Log) clear(index uint64) {
	log.deleteBefore(index-1)
}
func (log *Log) deleteBefore(index uint64) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	log.pause(true)
	defer log.pause(false)
	cutoff:=log.indexs.deleteBefore(index)
	if cutoff>0{
		log.node.storage.Copy(log.name,log.flushName,cutoff,log.size-cutoff)
		log.node.storage.Rename(log.name,log.tmpName)
		log.node.storage.Rename(log.flushName,log.name)
		log.node.storage.Rm(log.indexs.tmpName)
		log.node.storage.Rm(log.tmpName)
		lastLogIndex:=log.node.lastLogIndex
		if lastLogIndex>0{
			meta:=log.indexs.lookup(lastLogIndex)
			log.node.lastLogTerm=meta.Term
			log.size=meta.Position+meta.Offset
		}else {
			log.node.lastLogTerm=0
			log.size=0
		}
	}
	Tracef("Log.deleteBefore %s deleteBefore %d cutoff %d logSize %d",log.node.address,index,cutoff,log.size)

}
func (log *Log) copyAfter(index uint64,max int)(entries []*Entry) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	metas:=log.indexs.copyAfter(index,max)
	return log.batchRead(metas)
}
func (log *Log) copyRange(startIndex uint64,endIndex uint64) []*Entry{
	metas:=log.indexs.copyRange(startIndex,endIndex)
	//Tracef("Log.copyRange %s startIndex %d endIndex %d",log.node.address,metas[0].Index,metas[len(metas)-1].Index)
	return log.batchRead(metas)
}
func (log *Log) applyCommited() {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	lastLogIndex:=log.node.lastLogIndex
	if lastLogIndex==0{
		return
	}
	var startIndex =maxUint64(log.node.stateMachine.lastApplied+1,1)
	var endIndex =log.node.commitIndex.Id()
	if startIndex>endIndex{
		return
	}
	if endIndex-startIndex>DefaultMaxBatch{
		index:=startIndex
		for {
			log.applyCommitedRange(index,index+DefaultMaxBatch)
			index+=DefaultMaxBatch
			if endIndex-index<=DefaultMaxBatch{
				log.applyCommitedRange(index,endIndex)
				break
			}
		}
	}else {
		log.applyCommitedRange(startIndex,endIndex)
	}
}

func (log *Log) applyCommitedRange(startIndex uint64,endIndex uint64) {
	entries:=log.copyRange(startIndex,endIndex)
	if len(entries)==0{
		return
	}
	for i:=0;i<len(entries);i++{
		//Tracef("Log.applyCommitedRange %s Index %d Type %d",log.node.address,entries[i].Index,entries[i].CommandType)
		command:=log.node.commandType.clone(entries[i].CommandType)
		var err error
		if entries[i].CommandType>=0{
			err=log.node.raftCodec.Decode(entries[i].Command,command)
		}else {
			err=log.node.commandCodec.Decode(entries[i].Command,command)
		}
		if err==nil{
			log.node.stateMachine.Apply(entries[i].Index,command)
		}

	}
	log.putEmtyEntries(entries)
	//Tracef("Log.applyCommitedRange %s startIndex %d endIndex %d",log.node.address,startIndex,endIndex)
}

func (log *Log) appendEntries(entries []*Entry)bool {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	if !log.node.isLeader(){
		if !log.check(entries){
			return false
		}
	}
	data,metas,_:=log.Encode(entries,log.size)
	log.putEmtyEntries(entries)
	if !log.node.isLeader(){
		if !log.indexs.check(metas){
			return false
		}
	}
	if log.node.lastLogIndex!=metas[0].Index-1{
		return false
	}
	log.append(data)
	log.indexs.appendMetas(metas)
	log.node.lastLogIndex=metas[len(metas)-1].Index
	log.node.lastLogTerm=metas[len(metas)-1].Term
	log.indexs.putEmtyMetas(metas)
	return true
}
func (log *Log) read(meta *Meta)*Entry {
	if meta==nil{
		return nil
	}
	b:=make([]byte,meta.Offset)
	_,err:=log.node.storage.SeekRead(DefaultLog,meta.Position,b)
	if err!=nil{
		return nil
	}
	entry:=log.getEmtyEntry()
	err=log.node.raftCodec.Decode(b,entry)
	if err!=nil{
		Errorf("Log.Decode %s",string(b))
		return nil
	}
	return entry
}

func (log *Log) batchRead(metas []*Meta)[]*Entry {
	if len(metas)==0{
		return nil
	}
	cursor:=metas[0].Position
	offset:=metas[len(metas)-1].Position+metas[len(metas)-1].Offset
	b:=make([]byte,offset-cursor)
	if _,err:=log.node.storage.SeekRead(DefaultLog,cursor,b);err!=nil{
		return nil
	}
	if entries,_,err := log.Decode(b,metas,cursor); err == nil {
		return entries
	}
	return nil
}

func (log *Log) append(b []byte) {
	log.node.storage.SeekWrite(DefaultLog,log.size,b)
	log.size+=uint64(len(b))
}

func (log *Log) load() error {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	lastLogIndex:=log.node.lastLogIndex
	err:=log.indexs.load()
	if err!=nil{
		return err
	}
	if !log.node.storage.Exists(DefaultLog){
		return errors.New(DefaultLog+" file is not existed")
	}
	if log.node.lastLogIndex>0{
		meta:=log.indexs.lookup(log.node.lastLogIndex)
		log.node.lastLogTerm=meta.Term
		log.size=meta.Position+meta.Offset
		log.node.recoverLogIndex=log.node.lastLogIndex
		log.node.nextIndex=log.node.lastLogIndex+1
	}
	if log.node.lastLogIndex>lastLogIndex{
		Tracef("Log.recover %s lastLogIndex %d",log.node.address,lastLogIndex)
	}
	return nil
}

func (log *Log)Decode(data []byte,metas []*Meta,cursor uint64,)([]*Entry,uint64,error)  {
	length:=len(metas)
	entries:=make([]*Entry,0,length)
	var log_ret uint64
	for j:=0;j<length;j++{
		ret:=metas[j].Position-cursor
		offset:=metas[j].Offset
		func(){
			defer func() {
				if err := recover(); err != nil {
					Errorln("%s %d %d %d",err,len(data),ret,ret+offset)
				}
			}()
			b:=data[ret:ret+offset]
			entry:=log.getEmtyEntry()
			err:=log.node.raftCodec.Decode(b,entry)
			if err!=nil{
				Errorf("Log.Decode %d %d %d %s",cursor,ret,offset,string(b))
			}
			entries=append(entries, entry)
			log_ret=uint64(ret+offset)
		}()
	}
	return entries,cursor+log_ret,nil
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
		meta.Position=ret
		length:=uint64(len(b))
		meta.Offset=length
		ret+=length
		metas=append(metas,meta)
	}
	return data,metas,nil
}


func (log *Log) compaction() error {
	//log.mu.Lock()
	//defer log.mu.Unlock()
	//md5, err:=log.loadMd5()
	//if err==nil&&md5!=""{
	//	if md5==log.node.storage.MD5(DefaultLog){
	//		return nil
	//	}
	//}
	//var compactionEntries =make(map[string]*Entry)
	//var entries	=make([]*Entry,0)
	//for i:=len(log.entries)-1;i>=0;i--{
	//	entry:= log.entries[i]
	//	key:=string(int32ToBytes(entry.CommandType))+entry.CommandId
	//	if _,ok:=compactionEntries[key];!ok{
	//		entries=append(entries,entry)
	//		compactionEntries[key]=entry
	//	}
	//}
	//for i,j:= 0,len(entries)-1;i<j;i,j=i+1,j-1{
	//	entries[i], entries[j] = entries[j], entries[i]
	//}
	//log.entries=entries
	//
	//b,metas,_:=log.Encode(log.entries,0)
	//var lastLogSize ,_= log.node.storage.Size(DefaultLog)
	//log.node.storage.SafeOverWrite(DefaultLog,b)
	//log.indexs.metas=metas
	//log.indexs.save()
	//log.ret=uint64(len(b))
	//if len(log.entries)>0{
	//	log.node.lastLogIndex=log.entries[len(log.entries)-1].Index
	//	log.node.lastLogTerm=log.entries[len(log.entries)-1].Term
	//}
	//var logSize ,_= log.node.storage.Size(DefaultLog)
	//Tracef("Log.compaction %s LogSize %d==>%d",log.node.address,lastLogSize,logSize)
	//log.saveMd5()
	//Tracef("Log.compaction %s md5 %s==>%s",log.node.address,md5,log.node.storage.MD5(DefaultLog))
	return nil
}
func (log *Log) saveMd5() {
	md5:=log.node.storage.MD5(DefaultLog)
	if len(md5)>0{
		log.node.storage.OverWrite(DefaultMd5,[]byte(md5))
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
func (log *Log) Update()bool {
	if log.work{
		log.work=false
		defer func() {if err := recover(); err != nil {}}()
		defer func() {log.work=true}()
		log.batchMu.Lock()
		defer log.batchMu.Unlock()
		if log.isPaused(){
			return true
		}
		if len(log.readyEntries)>log.maxBatch{
			entries:=log.readyEntries[:log.maxBatch]
			log.readyEntries=log.readyEntries[log.maxBatch:]
			log.ticker(entries)
			return true
		}else  if len(log.readyEntries)>0{
			entries:=log.readyEntries[:]
			log.readyEntries=log.readyEntries[len(log.readyEntries):]
			log.ticker(entries)
			return true
		}
	}
	return false
}

func (log *Log) run()  {
	go func() {
		for entry := range log.entryChan {
			func(){
				log.batchMu.Lock()
				defer log.batchMu.Unlock()
				log.readyEntries=append(log.readyEntries, entry)
				if len(log.readyEntries)>=log.maxBatch&&!log.isPaused(){
					entries:=log.readyEntries[:]
					log.readyEntries=nil
					log.readyEntries=make([]*Entry,0)
					log.ticker(entries)
				}
			}()
		}
	}()
	for {
		select {
		case <-log.compactionTicker.C:
			func(){
				defer func() {if err := recover(); err != nil {}}()
				//log.compaction()
			}()
		case <-log.stop:
			close(log.stop)
			log.stop=nil
			goto endfor
		}
	}
endfor:
	close(log.entryChan)
	log.entryChan=nil
	log.compactionTicker.Stop()
	log.compactionTicker=nil
	log.finish<-true
}
func (log *Log) ticker(entries []*Entry) {
	log.appendEntries(entries)
}
func (log *Log)Stop()  {
	if log.stop==nil{
		return
	}
	log.stop<-true
	select {
	case <-log.finish:
		close(log.finish)
		log.finish=nil
	}
}