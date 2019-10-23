package raft

import (
	"io"
	"errors"
	"os"
	"time"
	"sync"
	"compress/gzip"
)

const (
)
type Snapshot interface {
	Save(context interface{},w io.Writer) (int, error)
	Recover(context interface{},r io.Reader) (int, error)
}

type SnapshotReadWriter struct {
	mut 				sync.RWMutex
	node				*Node
	work 				bool
	name 				string
	ret 				uint64
	read_ret 			uint64
	lastIncludedIndex 	*PersistentUint64
	lastIncludedTerm	*PersistentUint64
	lastTarIndex		*PersistentUint64
	length				uint64
	ticker				*time.Ticker
	stop 				chan bool
	finish 				chan bool
	tarWork 			bool
	done 				bool
}
func newSnapshotReadWriter(node *Node,name string)*SnapshotReadWriter  {
	s:=&SnapshotReadWriter{
		node:node,
		work:true,
		name:name,
		ret:0,
		read_ret:0,
		ticker:time.NewTicker(DefaultTarTick),
		lastIncludedIndex:newPersistentUint64(node,DefaultLastIncludedIndex,0),
		lastIncludedTerm:newPersistentUint64(node,DefaultLastIncludedTerm,0),
		lastTarIndex:newPersistentUint64(node,DefaultLastTarIndex,0),
		stop:make(chan bool,1),
		finish:make(chan bool,1),
		tarWork:true,
	}
	go s.run()
	return s
}

func (s *SnapshotReadWriter) Reset(lastIncludedIndex,lastIncludedTerm uint64) {
	s.lastIncludedIndex.Set(lastIncludedIndex)
	s.lastIncludedTerm.Set(lastIncludedTerm)
	s.node.storage.Truncate(s.name+DefaultTmp,0)
	s.ret=0
	s.read_ret=0
}
func (s *SnapshotReadWriter) Write(p []byte) (n int, err error){
	err=s.node.storage.SeekWrite(s.name+DefaultTmp,s.ret,p)
	if err!=nil{
		return 0,err
	}
	n=len(p)
	s.ret+=uint64(n)
	return n,nil
}
func (s *SnapshotReadWriter) Rename()error{
	return s.node.storage.Rename(s.name+DefaultTmp,s.name)
}
func (s *SnapshotReadWriter) Append(offset uint64,p []byte) (n int, err error){
	err=s.node.storage.SeekWrite(DefaultTarGz,offset,p)
	if err!=nil{
		return 0,err
	}
	return len(p),nil
}

func (s *SnapshotReadWriter)Read(p []byte) (n int, err error){
	n,err=s.node.storage.SeekRead(s.name,s.read_ret,p)
	if err!=nil{
		return n,err
	}
	s.read_ret+=uint64(n)
	return n,nil
}

func (s *SnapshotReadWriter) load() error {
	s.lastIncludedIndex.load()
	s.lastIncludedTerm.load()
	if !s.node.storage.Exists(s.name){
		return errors.New(s.name+" file is not existed")
	}
	return nil
}

func (s *SnapshotReadWriter) AppendFile(name string) error {
	var (
		size int64
		offsize int64
		err error
		buf []byte
		source *os.File
		n int
	)
	size,err=s.node.storage.Size(name)
	s.node.storage.AppendWrite(DefaultTar,uint64ToBytes(uint64(size)))
	if size>DefaultReadFileBufferSize{
		buf = make([]byte, DefaultReadFileBufferSize)
		source,err=s.node.storage.FileReader(name)
		if err!=nil{
			return err
		}
		defer source.Close()
		for {
			offsize+=DefaultReadFileBufferSize
			n, err = source.Read(buf)
			if err!=nil && err != io.EOF {
				return err
			}
			s.node.storage.AppendWrite(DefaultTar,buf[:n])
			if size-offsize<=DefaultReadFileBufferSize{
				n, err = source.Read(buf[:size-offsize])
				if err!=nil && err != io.EOF {
					return err
				}
				s.node.storage.AppendWrite(DefaultTar,buf[:n])
				break
			}
		}
	}else {
		buf, err = s.node.storage.Load(name)
		if err!=nil {
			return err
		}
		s.node.storage.AppendWrite(DefaultTar,buf)
	}
	return nil
}
func (s *SnapshotReadWriter) RecoverFile(source *os.File,name string) error {
	var (
		size uint64
		offsize uint64
		err error
		buf []byte
		n int
	)
	b:=make([]byte,8)
	_,err =source.Read(b)
	if err != nil {
		return err
	}
	size=bytesToUint64(b)
	s.node.storage.Truncate(name,0)
	if size>DefaultReadFileBufferSize{
		buf = make([]byte, DefaultReadFileBufferSize)
		for {
			n, err = source.Read(buf)
			if err!=nil && err != io.EOF {
				return err
			}
			s.node.storage.SeekWrite(name,offsize,buf[:n])
			offsize+=DefaultReadFileBufferSize
			if size-offsize<=DefaultReadFileBufferSize{
				n, err = source.Read(buf[:size-offsize])
				if err!=nil && err != io.EOF {
					return err
				}
				s.node.storage.SeekWrite(name,offsize,buf[:n])
				offsize+=uint64(n)
				break
			}
		}
	}else {
		buf = make([]byte, size)
		n, err = source.Read(buf)
		if err!=nil && err != io.EOF {
			return err
		}
		s.node.storage.OverWrite(name,buf)
	}
	return nil
}
func (s *SnapshotReadWriter) Tar()error{
	if s.canTar(){
		s.disableTar()
		err:=s.tar()
		s.enableTar()
		return err
	}
	return nil
}
func (s *SnapshotReadWriter) tar() error {
	if !s.node.storage.Exists(DefaultCommitIndex){
		s.node.commitIndex.save()
		return errors.New(DefaultCommitIndex+" file is not existed")
	}
	if !s.node.storage.Exists(DefaultLastIncludedIndex){
		s.lastIncludedIndex.save()
		return errors.New(DefaultLastIncludedIndex+" file is not existed")
	}
	if !s.node.storage.Exists(DefaultLastIncludedTerm){
		s.lastIncludedTerm.save()
		return errors.New(DefaultLastIncludedTerm+" file is not existed")
	}
	if !s.node.storage.Exists(DefaultSnapshot){
		return errors.New(DefaultSnapshot+" file is not existed")
	}
	if !s.node.storage.Exists(DefaultIndex){
		return errors.New(DefaultIndex+" file is not existed")
	}
	if !s.node.storage.Exists(DefaultLog){
		return errors.New(DefaultLog+" file is not existed")
	}
	s.done=false
	lastTarIndex:=s.lastTarIndex.Id()
	s.node.storage.Truncate(DefaultTar,0)
	s.AppendFile(DefaultLastTarIndex)
	s.AppendFile(DefaultCommitIndex)
	s.AppendFile(DefaultLastIncludedIndex)
	s.AppendFile(DefaultLastIncludedTerm)
	s.AppendFile(DefaultSnapshot)
	s.AppendFile(DefaultIndex)
	s.AppendFile(DefaultLog)
	s.gz()
	s.lastTarIndex.Set(s.lastIncludedIndex.Id())
	Tracef("SnapshotReadWriter.tar %s lastTarIndex %d==%d",s.node.address,lastTarIndex,s.lastTarIndex.Id())
	s.done=true
	s.node.storage.Rm(DefaultTar)
	return nil
}
func (s *SnapshotReadWriter) untar() error {
	if s.node.storage.IsEmpty(DefaultTarGz){
		return errors.New(DefaultTarGz+" file is empty")
	}
	if !s.done{
		return nil
	}
	s.ungz()
	source,err:=s.node.storage.FileReader(DefaultTar)
	if err!=nil{
		return err
	}
	defer source.Close()
	s.RecoverFile(source,DefaultLastTarIndex)
	s.RecoverFile(source,DefaultCommitIndex)
	s.RecoverFile(source,DefaultLastIncludedIndex)
	s.RecoverFile(source,DefaultLastIncludedTerm)
	s.RecoverFile(source,DefaultSnapshot+DefaultTmp)
	s.Rename()
	s.RecoverFile(source,DefaultIndex)
	s.RecoverFile(source,DefaultLog)
	if s.node.storage.Exists(DefaultTar){
		s.node.storage.Rm(DefaultTar)
	}
	if s.node.storage.Exists(DefaultTarGz){
		s.node.storage.Rm(DefaultTarGz)
	}
	if s.node.storage.Exists(DefaultLastTarIndex){
		s.node.storage.Rm(DefaultLastTarIndex)
	}
	return nil
}
func (s *SnapshotReadWriter) gz() error {
	dest,err:=s.node.storage.FileWriter(DefaultTarGz)
	if err!=nil{
		return err
	}
	defer dest.Close()
	source,err:=s.node.storage.FileReader(DefaultTar)
	if err!=nil{
		return err
	}
	defer source.Close()
	size,err:=s.node.storage.Size(DefaultTar)
	if err!=nil{
		return err
	}
	writer:=gzip.NewWriter(dest)
	defer writer.Close()
	buf := make([]byte, DefaultReadFileBufferSize)
	var offsize int64=0
	for {
		n, err := source.Read(buf)
		if err!=nil && err != io.EOF {
			return err
		}
		writer.Write(buf[:n])
		offsize+=DefaultReadFileBufferSize
		if size-offsize<=DefaultReadFileBufferSize{
			n, err = source.Read(buf)
			if err!=nil && err != io.EOF {
				return err
			}
			writer.Write(buf[:n])
			offsize+=int64(n)
			break
		}
	}
	writer.Flush()
	return nil
}
func (s *SnapshotReadWriter) ungz() error {
	dest,err:=s.node.storage.FileWriter(DefaultTar)
	if err!=nil{
		return err
	}
	defer dest.Close()
	source,err:=s.node.storage.FileReader(DefaultTarGz)
	if err!=nil{
		return err
	}
	defer source.Close()
	reader, err:= gzip.NewReader(source)
	defer reader.Close()
	if err != nil {
		return err
	}
	_,err=io.Copy(dest,reader)
	if err != nil&&err!=io.ErrUnexpectedEOF {
		return err
	}
	return nil
}
func (s *SnapshotReadWriter) clear() error {
	return s.node.storage.Truncate(DefaultTarGz,0)
}
func (s *SnapshotReadWriter)canTar() bool {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.tarWork
}
func (s *SnapshotReadWriter)disableTar() {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.tarWork=false
}
func (s *SnapshotReadWriter)enableTar() {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.tarWork=true
}

func (s *SnapshotReadWriter)run()  {
	go func() {
		for range s.ticker.C{
			if s.node.install()&&s.lastIncludedIndex.Id()>s.lastTarIndex.Id()&&s.node.isLeader(){
				s.Tar()
			}
		}
	}()
	select {
	case <-s.stop:
		close(s.stop)
		s.stop=nil
	}
	s.ticker.Stop()
	s.finish<-true
}
func (s *SnapshotReadWriter)Stop()  {
	if s.stop==nil{
		return
	}
	s.stop<-true
	select {
	case <-s.finish:
		close(s.finish)
		s.finish=nil
	}
}