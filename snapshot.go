// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"compress/gzip"
	"errors"
	"github.com/hslam/code"
	"io"
	"os"
	"sync"
	"time"
)

type Snapshot interface {
	Save(context interface{}, w io.Writer) (int, error)
	Recover(context interface{}, r io.Reader) (int, error)
}

type snapshotReadWriter struct {
	mut               sync.RWMutex
	node              *Node
	work              bool
	name              string
	tmpName           string
	flushName         string
	tarName           string
	tarGzName         string
	ret               uint64
	read_ret          uint64
	lastIncludedIndex *persistentUint64
	lastIncludedTerm  *persistentUint64
	lastTarIndex      *persistentUint64
	length            uint64
	ticker            *time.Ticker
	tarWork           bool
	done              bool
	gzip              bool
}

func newSnapshotReadWriter(node *Node, name string, gzip bool) *snapshotReadWriter {
	s := &snapshotReadWriter{
		node:              node,
		work:              true,
		name:              name,
		tmpName:           name + DefaultTmp,
		flushName:         name + DefaultFlush,
		tarName:           DefaultTar,
		tarGzName:         DefaultTarGz,
		ret:               0,
		read_ret:          0,
		ticker:            time.NewTicker(DefaultTarTick),
		lastIncludedIndex: newPersistentUint64(node, DefaultLastIncludedIndex, 0),
		lastIncludedTerm:  newPersistentUint64(node, DefaultLastIncludedTerm, 0),
		lastTarIndex:      newPersistentUint64(node, DefaultLastTarIndex, 0),
		tarWork:           true,
		gzip:              gzip,
	}
	go s.run()
	return s
}
func (s *snapshotReadWriter) Gz() bool {
	return s.gzip
}
func (s *snapshotReadWriter) Gzip(gz bool) {
	s.gzip = gz
}
func (s *snapshotReadWriter) FileName() string {
	name := s.tarName
	if s.gzip {
		name = s.tarGzName
	}
	return name
}
func (s *snapshotReadWriter) Reset(lastIncludedIndex, lastIncludedTerm uint64) {
	s.lastIncludedIndex.Set(lastIncludedIndex)
	s.lastIncludedTerm.Set(lastIncludedTerm)
	s.node.storage.Truncate(s.flushName, 0)
	s.ret = 0
	s.read_ret = 0
}
func (s *snapshotReadWriter) Write(p []byte) (n int, err error) {
	err = s.node.storage.SeekWrite(s.flushName, s.ret, p)
	if err != nil {
		return 0, err
	}
	n = len(p)
	s.ret += uint64(n)
	return n, nil
}
func (s *snapshotReadWriter) Rename() error {
	defer func() {
		if s.node.storage.Exists(s.tmpName) {
			s.node.storage.Rm(s.tmpName)
		}
	}()
	s.node.storage.Rename(s.name, s.tmpName)
	return s.node.storage.Rename(s.flushName, s.name)
}
func (s *snapshotReadWriter) Append(offset uint64, p []byte) (n int, err error) {
	err = s.node.storage.SeekWrite(s.FileName(), offset, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *snapshotReadWriter) Read(p []byte) (n int, err error) {
	n, err = s.node.storage.SeekRead(s.name, s.read_ret, p)
	if err != nil {
		return n, err
	}
	s.read_ret += uint64(n)
	return n, nil
}

func (s *snapshotReadWriter) load() error {
	s.lastIncludedIndex.load()
	s.lastIncludedTerm.load()
	if !s.node.storage.Exists(s.name) {
		return errors.New(s.name + " file is not existed")
	}
	return nil
}

func (s *snapshotReadWriter) AppendFile(name string) error {
	var (
		size    int64
		offsize int64
		err     error
		buf     []byte
		source  *os.File
		n       int
	)
	size, err = s.node.storage.Size(name)
	size_buf := make([]byte, 8)
	code.EncodeUint64(size_buf, uint64(size))
	s.node.storage.AppendWrite(DefaultTar, size_buf)
	if size > DefaultReadFileBufferSize {
		buf = make([]byte, DefaultReadFileBufferSize)
		source, err = s.node.storage.FileReader(name)
		if err != nil {
			return err
		}
		defer source.Close()
		for {
			offsize += DefaultReadFileBufferSize
			n, err = source.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			s.node.storage.AppendWrite(DefaultTar, buf[:n])
			if size-offsize <= DefaultReadFileBufferSize {
				n, err = source.Read(buf[:size-offsize])
				if err != nil && err != io.EOF {
					return err
				}
				s.node.storage.AppendWrite(DefaultTar, buf[:n])
				break
			}
		}
	} else {
		buf, err = s.node.storage.Load(name)
		if err != nil {
			return err
		}
		s.node.storage.AppendWrite(DefaultTar, buf)
	}
	return nil
}
func (s *snapshotReadWriter) RecoverFile(source *os.File, name string) error {
	var (
		size    uint64
		offsize uint64
		err     error
		buf     []byte
		n       int
	)
	b := make([]byte, 8)
	_, err = source.Read(b)
	if err != nil {
		return err
	}
	code.DecodeUint64(b, &size)
	s.node.storage.Truncate(name, 0)
	if size > DefaultReadFileBufferSize {
		buf = make([]byte, DefaultReadFileBufferSize)
		for {
			n, err = source.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			s.node.storage.SeekWrite(name, offsize, buf[:n])
			offsize += DefaultReadFileBufferSize
			if size-offsize <= DefaultReadFileBufferSize {
				n, err = source.Read(buf[:size-offsize])
				if err != nil && err != io.EOF {
					return err
				}
				s.node.storage.SeekWrite(name, offsize, buf[:n])
				offsize += uint64(n)
				break
			}
		}
	} else {
		buf = make([]byte, size)
		n, err = source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		s.node.storage.OverWrite(name, buf)
	}
	return nil
}
func (s *snapshotReadWriter) Tar() error {
	if s.canTar() {
		s.disableTar()
		defer s.enableTar()
		err := s.tar()
		return err
	}
	return nil
}
func (s *snapshotReadWriter) tar() error {
	if !s.node.storage.Exists(DefaultLastIncludedIndex) {
		s.lastIncludedIndex.save()
		return errors.New(DefaultLastIncludedIndex + " file is not existed")
	}
	if !s.node.storage.Exists(DefaultLastIncludedTerm) {
		s.lastIncludedTerm.save()
		return errors.New(DefaultLastIncludedTerm + " file is not existed")
	}
	if !s.node.storage.Exists(DefaultSnapshot) {
		return errors.New(DefaultSnapshot + " file is not existed")
	}
	//if !s.node.storage.Exists(DefaultIndex) {
	//	return errors.New(DefaultIndex + " file is not existed")
	//}
	//if !s.node.storage.Exists(DefaultLog) {
	//	return errors.New(DefaultLog + " file is not existed")
	//}
	s.done = false
	lastTarIndex := s.lastTarIndex.Id()
	s.lastTarIndex.Set(s.lastIncludedIndex.Id())
	s.node.storage.Truncate(DefaultTar, 0)
	s.AppendFile(DefaultLastTarIndex)
	s.AppendFile(DefaultLastIncludedIndex)
	s.AppendFile(DefaultLastIncludedTerm)
	s.AppendFile(DefaultSnapshot)
	//s.AppendFile(DefaultIndex)
	//s.AppendFile(DefaultLog)
	if s.gzip {
		s.gz()
	}
	Tracef("snapshotReadWriter.tar %s lastTarIndex %d==>%d", s.node.address, lastTarIndex, s.lastTarIndex.Id())
	s.done = true
	if s.gzip {
		s.node.storage.Rm(s.tarName)
	}
	return nil
}
func (s *snapshotReadWriter) untar() error {
	if s.node.storage.IsEmpty(s.FileName()) {
		return errors.New(s.FileName() + " file is empty")
	}
	if !s.done {
		return nil
	}
	if s.gzip {
		s.ungz()
	}
	source, err := s.node.storage.FileReader(DefaultTar)
	if err != nil {
		return err
	}
	defer source.Close()
	s.RecoverFile(source, DefaultLastTarIndex)
	s.RecoverFile(source, DefaultLastIncludedIndex)
	s.RecoverFile(source, DefaultLastIncludedTerm)
	s.RecoverFile(source, s.name)
	//s.RecoverFile(source, DefaultIndex)
	//s.RecoverFile(source, DefaultLog)
	if !s.node.isLeader() {
		if s.node.storage.Exists(DefaultTar) {
			s.node.storage.Rm(DefaultTar)
		}
		if s.node.storage.Exists(DefaultLastTarIndex) {
			s.node.storage.Rm(DefaultLastTarIndex)
		}
	}
	if s.gzip {
		if s.node.storage.Exists(DefaultTarGz) {
			s.node.storage.Rm(DefaultTarGz)
		}
	}
	return nil
}
func (s *snapshotReadWriter) gz() error {
	dest, err := s.node.storage.FileWriter(DefaultTarGz)
	if err != nil {
		return err
	}
	defer dest.Close()
	source, err := s.node.storage.FileReader(DefaultTar)
	if err != nil {
		return err
	}
	defer source.Close()
	size, err := s.node.storage.Size(DefaultTar)
	if err != nil {
		return err
	}
	writer := gzip.NewWriter(dest)
	defer writer.Close()
	buf := make([]byte, DefaultReadFileBufferSize)
	var offsize int64 = 0
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		writer.Write(buf[:n])
		offsize += DefaultReadFileBufferSize
		if size-offsize <= DefaultReadFileBufferSize {
			n, err = source.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			writer.Write(buf[:n])
			offsize += int64(n)
			break
		}
	}
	writer.Flush()
	return nil
}
func (s *snapshotReadWriter) ungz() error {
	dest, err := s.node.storage.FileWriter(DefaultTar)
	if err != nil {
		return err
	}
	defer dest.Close()
	source, err := s.node.storage.FileReader(DefaultTarGz)
	if err != nil {
		return err
	}
	defer source.Close()
	reader, err := gzip.NewReader(source)
	defer reader.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(dest, reader)
	if err != nil && err != io.ErrUnexpectedEOF {
		return err
	}
	return nil
}
func (s *snapshotReadWriter) clear() error {
	return s.node.storage.Truncate(s.FileName(), 0)
}
func (s *snapshotReadWriter) canTar() bool {
	s.mut.RLock()
	defer s.mut.RUnlock()
	return s.tarWork
}
func (s *snapshotReadWriter) disableTar() {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.tarWork = false
}
func (s *snapshotReadWriter) enableTar() {
	s.mut.Lock()
	defer s.mut.Unlock()
	s.tarWork = true
}

func (s *snapshotReadWriter) run() {
	for range s.ticker.C {
		if s.node.install() && s.lastIncludedIndex.Id() > s.lastTarIndex.Id() && s.node.isLeader() {
			s.Tar()
		}
	}
}
func (s *snapshotReadWriter) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
	}
}
