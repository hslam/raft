package raft


type PipelineCommand struct {
	invoker RaftCommand
	cbChan chan interface{}
	cbErrorChan chan error
}

func NewPipelineCommand(invoker RaftCommand) *PipelineCommand {
	c:= &PipelineCommand{
		invoker:invoker,
		cbChan:make(chan interface{},1),
		cbErrorChan:make(chan error,1),
	}
	return c
}

type PipelineCommandChan chan *PipelineCommand

type Pipeline struct {
	node					*Node
	pipelineCommandChan 	PipelineCommandChan
	replyChan 				PipelineCommandChan
	maxPipelineCommand		int
}

func NewPipeline(node *Node,maxPipelineCommand int) *Pipeline {
	pipeline:= &Pipeline{
		node :					node,
		pipelineCommandChan:make(PipelineCommandChan,maxPipelineCommand),
		replyChan:make(PipelineCommandChan,maxPipelineCommand),
		maxPipelineCommand:maxPipelineCommand,
	}
	go pipeline.run()
	return pipeline
}

func (pipeline *Pipeline)GetMaxPipelineCommand()int {
	return pipeline.maxPipelineCommand
}
func (pipeline *Pipeline)run()  {
	for p := range pipeline.pipelineCommandChan {
		data,_:=p.invoker.Encode()
		pipeline.node.logIndex+=1
		entry:=&Entry{
			Index:pipeline.node.logIndex,
			Term:pipeline.node.currentTerm.Id(),
			Command:data,
			CommandType:p.invoker.Type(),
			CommandId:p.invoker.UniqueID(),
		}
		pipeline.node.log.entryChan<-entry
		res,err:=p.invoker.Do(pipeline.node)
		p.cbChan<-res
		p.cbErrorChan<-err
	}
}