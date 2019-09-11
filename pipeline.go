package raft

import (
	"time"
)

type PipelineCommand struct {
	invoker RaftCommand
	reply 	chan interface{}
	err 	chan error
}

func NewPipelineCommand(invoker RaftCommand) *PipelineCommand {
	c:= &PipelineCommand{
		invoker:invoker,
		reply:make(chan interface{},1),
		err:make(chan error,1),
	}
	return c
}

type PipelineCommandChan chan *PipelineCommand

type Pipeline struct {
	node						*Node
	pipelineCommandChan 		PipelineCommandChan
	readyInvokerChan			PipelineCommandChan
	maxPipelineCommand			int
}

func NewPipeline(node *Node,maxPipelineCommand int) *Pipeline {
	pipeline:= &Pipeline{
		node :					node,
		pipelineCommandChan:make(PipelineCommandChan,maxPipelineCommand*2),
		readyInvokerChan:make(PipelineCommandChan,maxPipelineCommand*2),
		maxPipelineCommand:maxPipelineCommand,
	}
	go pipeline.run()
	return pipeline
}

func (pipeline *Pipeline)GetMaxPipelineCommand()int {
	return pipeline.maxPipelineCommand
}
func (pipeline *Pipeline)run()  {
	go func() {
		for p := range pipeline.readyInvokerChan {
			for{
				if pipeline.node.commitIndex>0&&p.invoker.Index()<=pipeline.node.commitIndex{
					//var lastApplied  = pipeline.node.stateMachine.lastApplied
					reply,err:=pipeline.node.stateMachine.Apply(p.invoker.Index(),p.invoker)
					p.reply<-reply
					p.err<-err
					//Tracef("Pipeline.run %s lastApplied %d==>%d",pipeline.node.address,lastApplied,pipeline.node.stateMachine.lastApplied)
					goto endfor
				}
				time.Sleep(time.Microsecond*100)
			}
			endfor:
		}
	}()
	for p := range pipeline.pipelineCommandChan {
		data,_:=p.invoker.Encode()
		p.invoker.SetIndex(pipeline.node.nextIndex)
		pipeline.node.nextIndex+=1
		entry:=&Entry{
			Index:p.invoker.Index(),
			Term:pipeline.node.currentTerm.Id(),
			Command:data,
			CommandType:p.invoker.Type(),
			CommandId:p.invoker.UniqueID(),
		}
		pipeline.readyInvokerChan<-p
		pipeline.node.log.entryChan<-entry

	}
}