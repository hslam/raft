package raft

import(
	"errors"
	"fmt"
)

type RaftCommand interface {
	Private()bool
	Type()uint32
	UniqueID()string
	Do(node *Node)(interface{},error)
	Encode()([]byte,error)
	Decode(data []byte)(error)
}


type Invoker struct {
	codec 		Codec
	cmd 		Command
	private 	bool
}
func newInvoker(cmd Command,private bool,codec Codec) RaftCommand {
	return &Invoker{
		codec:codec,
		cmd: cmd,
		private:private,
	}
}
func (i *Invoker) Private()bool{
	return i.private
}
func (i Invoker) Type() uint32{
	return i.cmd.Type()
}
func (i Invoker) UniqueID()string{
	return i.cmd.UniqueID()
}
func (i Invoker) Do(node *Node)(interface{},error) {
	return i.cmd.Do(node)
}
func (i Invoker) Encode()([]byte,error) {
	return i.codec.Encode(i.cmd)
}
func (i Invoker) Decode(data []byte)(error) {
	return i.codec.Decode(data,i.cmd)
}

type Commands struct {
	cmds  []Command
}
func newCommands() *Commands {
	return &Commands{
		cmds:make([]Command,0),
	}
}
func (p *Commands) AddCommand(cmd Command)error {
	if cmd == nil {
		return errors.New("Command can not be nil")
	}
	p.cmds= append(p.cmds,cmd)
	return nil
}
func (p Commands) Do(node *Node) {
	for _, item := range p.cmds {
		item.Do(node)
	}
}


type CommandMap struct {
	cmdMap map[string]bool
	cmds  []RaftCommand
}

func (m *CommandMap) AddCommand(cmd RaftCommand) error{
	if cmd == nil {
		return errors.New("Command can not be nil")
	}
	if _,ok:=m.cmdMap[cmd.UniqueID()];!ok{
		return errors.New(fmt.Sprintf("Command %s is existed", cmd.UniqueID()))
	}
	m.cmdMap[cmd.UniqueID()] = true
	m.cmds= append(m.cmds,cmd)
	return nil
}

func (p CommandMap) Do(node *Node) {
	for _, item := range p.cmds {
		item.Do(node)
	}
}

