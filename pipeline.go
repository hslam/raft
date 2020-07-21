package raft

import (
	"sync"
	"time"
)

var (
	pipelineCommandPool *sync.Pool
)

func init() {
	pipelineCommandPool = &sync.Pool{
		New: func() interface{} {
			return &PipelineCommand{}
		},
	}
}

type PipelineCommand struct {
	invoker RaftCommand
	reply   chan interface{}
	err     chan error
}

func NewPipelineCommand(invoker RaftCommand, reply chan interface{}, err chan error) *PipelineCommand {
	c := pipelineCommandPool.Get().(*PipelineCommand)
	c.invoker = invoker
	c.reply = reply
	c.err = err
	return c
}

type PipelineCommandChan chan *PipelineCommand

type Pipeline struct {
	node                *Node
	pipelineCommandChan PipelineCommandChan
	readyInvokerChan    PipelineCommandChan
	maxPipelineCommand  int
}

func NewPipeline(node *Node, maxPipelineCommand int) *Pipeline {
	pipeline := &Pipeline{
		node:                node,
		pipelineCommandChan: make(PipelineCommandChan, maxPipelineCommand*2),
		readyInvokerChan:    make(PipelineCommandChan, maxPipelineCommand*2),
		maxPipelineCommand:  maxPipelineCommand,
	}
	go pipeline.run()
	return pipeline
}

func (pipeline *Pipeline) GetMaxPipelineCommand() int {
	return pipeline.maxPipelineCommand
}
func (pipeline *Pipeline) run() {
	go func() {
		for p := range pipeline.readyInvokerChan {
			for {
				if pipeline.node.commitIndex.Id() > 0 && p.invoker.Index() <= pipeline.node.commitIndex.Id() {
					//var lastApplied  = pipeline.node.stateMachine.lastApplied
					reply, err := pipeline.node.stateMachine.Apply(p.invoker.Index(), p.invoker)
					func() {
						defer func() {
							if err := recover(); err != nil {
							}
						}()
						p.reply <- reply
						p.err <- err
					}()
					//Tracef("Pipeline.run %s lastApplied %d==>%d",pipeline.node.address,lastApplied,pipeline.node.stateMachine.lastApplied)
					goto endfor
				}
				time.Sleep(time.Microsecond * 100)
			}
		endfor:
			invoker := p.invoker
			invokerPool.Put(invoker)
			p.err = nil
			p.reply = nil
			p.invoker = nil
			pipelineCommandPool.Put(p)
		}
	}()
	for p := range pipeline.pipelineCommandChan {
		data, _ := p.invoker.Encode()
		p.invoker.SetIndex(pipeline.node.nextIndex)
		pipeline.node.nextIndex += 1
		entry := pipeline.node.log.getEmtyEntry()
		entry.Index = p.invoker.Index()
		entry.Term = pipeline.node.currentTerm.Id()
		entry.Command = data
		entry.CommandType = p.invoker.Type()
		entry.CommandId = p.invoker.UniqueID()
		pipeline.readyInvokerChan <- p
		pipeline.node.log.entryChan <- entry

	}
}
