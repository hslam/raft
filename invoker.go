package raft
import (
	"sync"
)
var (
	invokerPool			*sync.Pool
)
func init() {
	invokerPool= &sync.Pool{
		New: func() interface{} {
			return &Invoker{}
		},
	}
}

type RaftCommand interface {
	Type()int32
	UniqueID()string
	Do(Context interface{})(interface{},error)
	Private()bool
	SetIndex(index uint64)
	Index()uint64
	Encode()([]byte,error)
	Decode(data []byte)(error)
}

type Invoker struct {
	codec 		Codec
	cmd 		Command
	private 	bool
	index 		uint64
}
func newInvoker(cmd Command,private bool,codec Codec) RaftCommand {
	i:=invokerPool.Get().(*Invoker)
	i.codec=codec
	i.cmd=cmd
	i.private=private
	return &Invoker{
		codec:codec,
		cmd: cmd,
		private:private,
	}
}
func (i *Invoker) Private()bool{
	return i.private
}
func (i *Invoker) SetIndex(index uint64){
	i.index=index
}
func (i *Invoker) Index() uint64{
	return i.index
}
func (i Invoker) Type() int32{
	return i.cmd.Type()
}
func (i Invoker) UniqueID()string{
	return i.cmd.UniqueID()
}
func (i Invoker) Do(Context interface{})(interface{},error) {
	return i.cmd.Do(Context)
}
func (i Invoker) Encode()([]byte,error) {
	return i.codec.Encode(i.cmd)
}
func (i Invoker) Decode(data []byte)(error) {
	return i.codec.Decode(data,i.cmd)
}
