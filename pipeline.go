package raft

import (
	"time"
)

type PipelineCommandChan chan *Invoker

type Pipeline struct {
	node                *Node
	pipelineCommandChan PipelineCommandChan
	readyInvokerChan    PipelineCommandChan
	maxPipelineCommand  int
	buffer              []byte
}

func NewPipeline(node *Node, maxPipelineCommand int) *Pipeline {
	pipeline := &Pipeline{
		node:                node,
		pipelineCommandChan: make(PipelineCommandChan, maxPipelineCommand*2),
		readyInvokerChan:    make(PipelineCommandChan, maxPipelineCommand*2),
		maxPipelineCommand:  maxPipelineCommand,
		buffer:              make([]byte, 1024*64),
	}
	go pipeline.run()
	return pipeline
}

func (pipeline *Pipeline) GetMaxPipelineCommand() int {
	return pipeline.maxPipelineCommand
}
func (pipeline *Pipeline) run() {
	go func() {
		for invoker := range pipeline.readyInvokerChan {
			for {
				if pipeline.node.commitIndex > 0 && invoker.index <= pipeline.node.commitIndex {
					//var lastApplied  = pipeline.node.stateMachine.lastApplied
					var err error
					invoker.Reply, invoker.Error, err = pipeline.node.stateMachine.Apply(invoker.index, invoker.Command)
					if err != nil {
						time.Sleep(time.Microsecond * 100)
						continue
					}
					invoker.done()
					//Tracef("Pipeline.run %s lastApplied %d==>%d",pipeline.node.address,lastApplied,pipeline.node.stateMachine.lastApplied)
					goto endfor
				}
				time.Sleep(time.Microsecond * 100)
			}
		endfor:
		}
	}()
	for invoker := range pipeline.pipelineCommandChan {
		invoker.index = pipeline.node.nextIndex
		pipeline.node.nextIndex += 1
		var data []byte
		if invoker.Command.Type() >= 0 {
			b, _ := pipeline.node.codec.Marshal(pipeline.buffer, invoker.Command)
			data = make([]byte, len(b))
			copy(data, b)
		} else {
			b, _ := pipeline.node.raftCodec.Marshal(pipeline.buffer, invoker.Command)
			data = make([]byte, len(b))
			copy(data, b)
		}
		entry := pipeline.node.log.getEmtyEntry()
		entry.Index = invoker.index
		entry.Term = pipeline.node.currentTerm.Id()
		entry.Command = data
		entry.CommandType = invoker.Command.Type()
		entry.CommandId = invoker.Command.UniqueID()
		pipeline.readyInvokerChan <- invoker
		pipeline.node.log.entryChan <- entry

	}
}
