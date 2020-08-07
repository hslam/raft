package raft

import (
	"errors"
	"github.com/hslam/wal"
	"sync"
	"time"
)

type Log struct {
	mu               sync.Mutex
	batchMu          sync.Mutex
	pauseMu          sync.Mutex
	node             *Node
	wal              *wal.Log
	entryChan        chan *Entry
	readyEntries     []*Entry
	maxBatch         int
	compactionTicker *time.Ticker
	buf              []byte
	entryPool        *sync.Pool
	stop             chan bool
	finish           chan bool
	work             bool
	paused           bool
}

func newLog(node *Node) *Log {
	log := &Log{
		node:             node,
		entryChan:        make(chan *Entry, DefaultMaxCacheEntries),
		readyEntries:     make([]*Entry, 0),
		maxBatch:         DefaultMaxBatch,
		compactionTicker: time.NewTicker(DefaultCompactionTick),
		buf:              make([]byte, 1024*64),
		stop:             make(chan bool, 1),
		finish:           make(chan bool, 1),
		work:             true,
	}
	log.wal, _ = wal.Open(node.storage.data_dir)
	log.entryPool = &sync.Pool{
		New: func() interface{} {
			return &Entry{}
		},
	}
	go log.run()
	return log
}
func (log *Log) getEmtyEntry() *Entry {
	return log.entryPool.Get().(*Entry)
}
func (log *Log) putEmtyEntry(entry *Entry) {
	entry.Index = 0
	entry.Term = 0
	entry.Command = []byte{}
	entry.CommandType = 0
	entry.CommandId = ""
	log.entryPool.Put(entry)
}
func (log *Log) putEmtyEntries(entries []*Entry) {
	for _, entry := range entries {
		log.putEmtyEntry(entry)
	}
}
func (log *Log) Lock() {
	log.mu.Lock()
}
func (log *Log) Unlock() {
	log.mu.Unlock()
}
func (log *Log) pause(p bool) {
	log.pauseMu.Lock()
	defer log.pauseMu.Unlock()
	log.paused = p
}
func (log *Log) isPaused() bool {
	log.pauseMu.Lock()
	defer log.pauseMu.Unlock()
	return log.paused
}
func (log *Log) checkPaused() {
	for {
		if !log.isPaused() {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
}
func (log *Log) checkIndex(index uint64) bool {
	if ok, err := log.wal.IsExist(index); err != nil {
		return false
	} else {
		return ok
	}
}

func (log *Log) lookup(index uint64) *Entry {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	return log.read(index)
}

func (log *Log) consistencyCheck(index uint64, term uint64) (ok bool) {
	if index == 0 {
		return false
	}
	if index > log.node.lastLogIndex {
		return false
	}
	if index == log.node.lastLogIndex {
		if term == log.node.lastLogTerm {
			return true
		}
	}
	entry := log.node.log.lookup(index)
	if entry == nil {
		return false
	}
	if entry.Term != term {
		log.node.log.deleteAfter(index)
		log.node.nextIndex = log.node.lastLogIndex + 1
		return false
	}
	return true
}
func (log *Log) check(entries []*Entry) bool {
	lastIndex := entries[0].Index
	for i := 1; i < len(entries); i++ {
		if entries[i].Index == lastIndex+1 {
			lastIndex = entries[i].Index
		} else {
			return false
		}
	}
	return true
}

func (log *Log) deleteAfter(index uint64) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.pause(true)
	defer log.pause(false)
	if index == log.node.firstLogIndex {
		log.wal.Reset()
		log.wal.InitFirstIndex(index)
		return
	}
	log.node.lastLogIndex = index - 1
	log.wal.Truncate(index - 1)
	lastLogIndex := log.node.lastLogIndex
	if lastLogIndex > 0 {
		entry := log.read(lastLogIndex)
		log.node.lastLogTerm = entry.Term
	} else {
		log.node.lastLogTerm = 0
	}
}

func (log *Log) deleteBefore(index uint64) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	log.pause(true)
	defer log.pause(false)
	log.wal.Clean(index)
	log.node.firstLogIndex, _ = log.wal.FirstIndex()
	Tracef("Log.deleteBefore %s deleteBefore %d", log.node.address, index)
}
func (log *Log) startIndex(index uint64) uint64 {
	return maxUint64(log.node.firstLogIndex, index)
}
func (log *Log) endIndex(index uint64) uint64 {
	return minUint64(index, log.node.lastLogIndex)
}
func (log *Log) copyAfter(index uint64, max int) (entries []*Entry) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	startIndex := log.startIndex(index)
	endIndex := log.endIndex(startIndex + uint64(max))
	return log.copyRange(startIndex, endIndex)
}
func (log *Log) copyRange(startIndex uint64, endIndex uint64) []*Entry {
	//Tracef("Log.copyRange %s startIndex %d endIndex %d",log.node.address,metas[0].Index,metas[len(metas)-1].Index)
	return log.batchRead(startIndex, endIndex)
}

func (log *Log) applyCommited() {
	log.node.stateMachine.Lock()
	defer log.node.stateMachine.Unlock()
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	lastLogIndex := log.node.lastLogIndex
	if lastLogIndex == 0 {
		return
	}
	var startIndex = maxUint64(log.node.stateMachine.lastApplied+1, 1)
	var endIndex = log.node.commitIndex
	if startIndex > endIndex {
		return
	}
	if endIndex-startIndex > DefaultMaxBatch {
		index := startIndex
		for {
			log.applyCommitedRange(index, index+DefaultMaxBatch)
			index += DefaultMaxBatch
			if endIndex-index <= DefaultMaxBatch {
				log.applyCommitedRange(index, endIndex)
				break
			}
		}
	} else {
		log.applyCommitedRange(startIndex, endIndex)
	}
}

func (log *Log) applyCommitedEnd(endIndex uint64) {
	log.node.stateMachine.Lock()
	defer log.node.stateMachine.Unlock()
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	lastLogIndex := log.node.lastLogIndex
	if lastLogIndex == 0 {
		return
	}
	var startIndex = maxUint64(log.node.stateMachine.lastApplied+1, 1)
	endIndex = minUint64(log.node.commitIndex, endIndex)
	if startIndex > endIndex {
		return
	}
	if endIndex-startIndex > DefaultMaxBatch {
		index := startIndex
		for {
			log.applyCommitedRange(index, index+DefaultMaxBatch)
			index += DefaultMaxBatch
			if endIndex-index <= DefaultMaxBatch {
				log.applyCommitedRange(index, endIndex)
				break
			}
		}
	} else {
		log.applyCommitedRange(startIndex, endIndex)
	}
}

func (log *Log) applyCommitedRange(startIndex uint64, endIndex uint64) {
	//Tracef("Log.applyCommitedRange %s startIndex %d endIndex %d Start",log.node.address,startIndex,endIndex)
	entries := log.copyRange(startIndex, endIndex)
	if entries == nil || len(entries) == 0 {
		return
	}
	//Tracef("Log.applyCommitedRange %s startIndex %d endIndex %d length %d",log.node.address,startIndex,endIndex,len(entries))
	for i := 0; i < len(entries); i++ {
		//Tracef("Log.applyCommitedRange %s Index %d Type %d",log.node.address,entries[i].Index,entries[i].CommandType)
		command := log.node.commandType.clone(entries[i].CommandType)
		var err error
		if entries[i].CommandType >= 0 {
			err = log.node.codec.Unmarshal(entries[i].Command, command)
		} else {
			err = log.node.raftCodec.Unmarshal(entries[i].Command, command)
		}
		if err == nil {
			log.node.stateMachine.apply(entries[i].Index, command)
		} else {
			Errorf("Log.applyCommitedRange %s %d error %s", log.node.address, i, err)
		}
		log.node.commandType.put(command)
	}
	log.putEmtyEntries(entries)
	//Tracef("Log.applyCommitedRange %s startIndex %d endIndex %d End %d",log.node.address,startIndex,endIndex,len(entries))
}

func (log *Log) appendEntries(entries []*Entry) bool {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	if !log.node.isLeader() {
		if !log.check(entries) {
			return false
		}
	}
	if log.node.lastLogIndex != entries[0].Index-1 {
		return false
	}
	log.Write(entries)
	log.node.lastLogIndex = entries[len(entries)-1].Index
	log.node.lastLogTerm = entries[len(entries)-1].Term
	log.putEmtyEntries(entries)
	return true
}
func (log *Log) read(index uint64) *Entry {
	b, err := log.wal.Read(index)
	if err != nil {
		return nil
	}
	entry := log.getEmtyEntry()
	err = log.node.raftCodec.Unmarshal(b, entry)
	if err != nil {
		Errorf("Log.Decode %s", string(b))
		return nil
	}
	return entry
}

func (log *Log) batchRead(startIndex uint64, endIndex uint64) []*Entry {
	entries := make([]*Entry, 0, endIndex-startIndex+1)
	for i := startIndex; i < endIndex+1; i++ {
		entries = append(entries, log.read(i))
	}
	return entries
}

func (log *Log) load() (err error) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.checkPaused()
	lastLogIndex := log.node.lastLogIndex

	log.node.firstLogIndex, err = log.wal.FirstIndex()
	if err != nil {
		return err
	}
	log.node.lastLogIndex, err = log.wal.LastIndex()
	if err != nil {
		return err
	}
	if log.node.lastLogIndex > 0 {
		entry := log.read(log.node.lastLogIndex)
		log.node.lastLogTerm = entry.Term
		log.node.recoverLogIndex = log.node.lastLogIndex
		log.node.nextIndex = log.node.lastLogIndex + 1
	}
	if log.node.lastLogIndex > lastLogIndex {
		Tracef("Log.recover %s lastLogIndex %d", log.node.address, lastLogIndex)
	}
	return nil
}

func (log *Log) Write(entries []*Entry) (err error) {
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		b, err := log.node.raftCodec.Marshal(log.buf, entry)
		if err != nil {
			return err
		}
		if err = log.wal.Write(entry.Index, b); err != nil {
			return err
		}
	}
	Tracef("Log.Write %d", len(entries))
	return log.wal.Flush()
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
	//md5 := log.node.storage.MD5(DefaultLog)
	//if len(md5) > 0 {
	//	log.node.storage.OverWrite(DefaultMd5, []byte(md5))
	//}
}

func (log *Log) loadMd5() (string, error) {
	if !log.node.storage.Exists(DefaultMd5) {
		return "", errors.New("md5 file is not existed")
	}
	b, err := log.node.storage.Load(DefaultMd5)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
func (log *Log) Update() bool {
	if log.work {
		log.work = false
		defer func() {
			if err := recover(); err != nil {
			}
		}()
		defer func() { log.work = true }()
		log.batchMu.Lock()
		defer log.batchMu.Unlock()
		if log.isPaused() {
			return true
		}
		if len(log.readyEntries) > log.maxBatch {
			entries := log.readyEntries[:log.maxBatch]
			log.readyEntries = log.readyEntries[log.maxBatch:]
			log.ticker(entries)
			return true
		} else if len(log.readyEntries) > 0 {
			entries := log.readyEntries[:]
			log.readyEntries = log.readyEntries[len(log.readyEntries):]
			log.ticker(entries)
			return true
		}
	}
	return false
}

func (log *Log) run() {
	go func() {
		for entry := range log.entryChan {
			func() {
				log.batchMu.Lock()
				defer log.batchMu.Unlock()
				log.readyEntries = append(log.readyEntries, entry)
				if len(log.readyEntries) >= log.maxBatch && !log.isPaused() {
					entries := log.readyEntries[:]
					log.readyEntries = nil
					log.readyEntries = make([]*Entry, 0)
					log.ticker(entries)
				}
			}()
		}
	}()
	for {
		select {
		case <-log.compactionTicker.C:
			func() {
				defer func() {
					if err := recover(); err != nil {
					}
				}()
				//log.compaction()
			}()
		case <-log.stop:
			close(log.stop)
			log.stop = nil
			goto endfor
		}
	}
endfor:
	close(log.entryChan)
	log.entryChan = nil
	log.compactionTicker.Stop()
	log.compactionTicker = nil
	log.finish <- true
}
func (log *Log) ticker(entries []*Entry) {
	log.appendEntries(entries)
}
func (log *Log) Stop() {
	if log.stop == nil {
		return
	}
	log.stop <- true
	select {
	case <-log.finish:
		close(log.finish)
		log.finish = nil
	}
}
