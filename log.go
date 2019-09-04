package raft

import (
	"time"
	"sync"
)

type Log struct {
	mu 						sync.Mutex
	batchMu 				sync.Mutex
	node					*Node
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
		entryChan:				make(chan *Entry,DefaultMaxCacheEntries),
		readyEntries:			make([]*Entry,0),
		maxBatch:				DefaultMaxBatch,
		appendEntriesTicker:	time.NewTicker(DefaultMaxDelay),
		compactionTicker:	time.NewTicker(DefaultCompactionTick),
	}
	go log.run()
	return log
}
func (log *Log) append(entry *Entry) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.entries=append(log.entries,entry)
	log.save()
}

func (log *Log) appendEntries(entries []*Entry) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.entries=append(log.entries,entries...)
	log.save()
}
func (log *Log) recovery() error {
	log.mu.Lock()
	defer log.mu.Unlock()
	if !log.node.storage.Exists(DefaultLog){
		return nil
	}
	b, err := log.node.storage.Load(DefaultLog)
	if err != nil {
		return nil
	}
	logStorage := &LogStorage{}
	if err = log.node.codec.Decode(b, logStorage); err != nil {
		return err
	}
	log.node.votedFor=logStorage.VotedFor
	log.entries=logStorage.Entries
	for _,entry:=range log.entries{
		command:=log.node.commandType.clone(entry.CommandType)
		err:=log.node.codec.Decode(entry.Command,command)
		if err==nil{
			log.node.stateMachine.Apply(command)
		}
	}
	if len(log.entries)>0{
		log.node.lastLogIndex=log.entries[len(log.entries)-1].Index
		log.node.lastLogTerm=log.entries[len(log.entries)-1].Term
		log.node.logIndex=log.node.lastLogIndex
	}
	return nil
}
func (log *Log) compaction() error {
	log.mu.Lock()
	defer log.mu.Unlock()
	var compactionEntries =make(map[string]*Entry)
	var entries	=make([]*Entry,0)
	for i:=len(log.entries)-1;i>=0;i--{
		entry:= log.entries[i]
		if _,ok:=compactionEntries[entry.CommandId];!ok{
			entries=append(entries,entry)
			compactionEntries[entry.CommandId]=entry
		}
	}
	for i,j:= 0,len(entries)-1;i<j;i,j=i+1,j-1{
		entries[i], entries[j] = entries[j], entries[i]
	}
	log.entries=entries
	logStorage := &LogStorage{
		VotedFor:log.node.votedFor,
		Entries:log.entries,
	}
	b, _ := log.node.codec.Encode(logStorage)
	log.node.storage.SafeOverWrite(DefaultLog,b)
	return nil
}
func (log *Log) save() {
	//log.mu.Lock()
	//defer log.mu.Unlock()

	logStorage := &LogStorage{
		VotedFor:log.node.votedFor,
		Entries:log.entries,
	}
	b, _ := log.node.codec.Encode(logStorage)
	log.node.storage.SafeOverWrite(DefaultLog,b)
	if len(log.entries)>0{
		log.node.lastLogIndex=log.entries[len(log.entries)-1].Index
		log.node.lastLogTerm=log.entries[len(log.entries)-1].Term
		log.node.logIndex=log.node.lastLogIndex
	}
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
			log.compaction()
		}
	}
}
func (log *Log) ticker(entries []*Entry) {
	log.appendEntries(entries)
	log.node.appendEntries(entries)
}