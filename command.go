package raft

import(
	"errors"
"fmt"
)
type Command interface {
	Name()string
	Do()

}

type Invoker struct {
	cmd Command
}

func (p *Invoker) SetCommand(cmd Command) error{
	if cmd == nil {
		return errors.New("Command can not be nil")
	}
	p.cmd = cmd
	return nil
}
func (p Invoker) Name() string{
	return ""
}
func (p Invoker) Do() {
	p.cmd.Do()
}


type Commands struct {
	cmds  []Command
}

func (p *Commands) AddCommand(cmd Command)error {
	if cmd == nil {
		return errors.New("Command can not be nil")
	}
	p.cmds= append(p.cmds,cmd)
	return nil
}
func (p Commands) Name() string{
	return ""
}
func (p Commands) Do() {
	for _, item := range p.cmds {
		item.Do()
	}
}


type CommandMap struct {
	cmdmap map[string]Command
}

func (p *CommandMap) RegisterCommand(cmd Command) error{
	if cmd == nil {
		return errors.New("Command can not be nil")
	} else if p.cmdmap[cmd.Name()] != nil {
		return errors.New(fmt.Sprintf("CommandMap.RegisterCommand: %s", cmd.Name()))
	}
	p.cmdmap[cmd.Name()] = cmd
	return nil
}

func (p CommandMap) Name() string{
	return ""
}
func (p CommandMap) Do() {
	for _, item := range p.cmdmap {
		item.Do()
	}
}
