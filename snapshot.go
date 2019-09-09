package raft

type Snapshot interface {
	Save(context interface{}) ([]byte, error)
	Recover(context interface{},snapshot []byte) error
}