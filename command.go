package raft

import (
	"reflect"
)

type Command interface {
	Type()uint32
	UniqueID()string
	Do(node *Node)(interface{},error)
}


type CommandType struct {
	types map[uint32]Command
}

func (m *CommandType) register(cmd Command) error{
	if _,ok:=m.types[cmd.Type()];ok{
		return ErrCommandTypeExisted
	}
	m.types[cmd.Type()] = cmd
	return nil
}
func (m *CommandType) clone(Type uint32)Command{
	if command,ok:=m.types[Type];ok{
		command_copy := reflect.New(reflect.Indirect(reflect.ValueOf(command)).Type()).Interface()
		return command_copy.(Command)
	}
	Debugf("CommandType.clone Unregistered %d",Type)
	return nil
}
func (m *CommandType)exists(cmd Command) bool{
	_,ok:=m.types[cmd.Type()];
	return ok
}
