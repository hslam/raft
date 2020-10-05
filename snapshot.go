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

// Snapshot saves a snapshot and recovers from a snapshot.
type Snapshot interface {
	Save(context interface{}, w io.Writer) (int, error)
	Recover(context interface{}, r io.Reader) (int, error)
}

type snapshotReadWriter struct {
	mut               sync.RWMutex
	node              *node
	work              bool
	name              string
	tmpName           string
	flushName         string
	tarName           string
	tarGzName         string
	ret               uint64
	readRet           uint64
	lastIncludedIndex *persistentUint64
	lastIncludedTerm  *persistentUint64
	lastTarIndex      *persistentUint64
	length            uint64
	ticker            *time.Ticker
	tarWork           bool
	done              bool
	gzip              bool
}

func newSnapshotReadWriter(n *node, name string, gzip bool) *snapshotReadWriter {
	s := &snapshotReadWriter{
		node:              n,
		work:              true,
		name:              name,
		tmpName:           name + defaultTmp,
		flushName:         name + defaultFlush,
		tarName:           defaultTar,
		tarGzName:         defaultTarGz,
		ret:               0,
		readRet:           0,
		ticker:            time.NewTicker(defaultTarTick),
		lastIncludedIndex: newPersistentUint64(n, defaultLastIncludedIndex, 0, 0),
		lastIncludedTerm:  newPersistentUint64(n, defaultLastIncludedTerm, 0, 0),
		lastTarIndex:      newPersistentUint64(n, defaultLastTarIndex, 0, 0),
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
	s.readRet = 0
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
	n, err = s.node.storage.SeekRead(s.name, s.readRet, p)
	if err != nil {
		return n, err
	}
	s.readRet += uint64(n)
	return n, nil
}

func (s *snapshotReadWriter) load() error {
	s.lastIncludedIndex.load()
	s.lastIncludedTerm.load()
	if !s.node.storage.Exists(s.name) {
		if s.node.storage.Exists(s.tmpName) {
			return s.node.storage.Rename(s.tmpName, s.name)
		}
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
	if err != nil {
		return err
	}
	sizeBuf := make([]byte, 8)
	code.EncodeUint64(sizeBuf, uint64(size))
	s.node.storage.AppendWrite(defaultTar, sizeBuf)
	if size > defaultReadFileBufferSize {
		buf = make([]byte, defaultReadFileBufferSize)
		source, err = s.node.storage.FileReader(name)
		if err != nil {
			return err
		}
		defer source.Close()
		for {
			offsize += defaultReadFileBufferSize
			n, err = source.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			s.node.storage.AppendWrite(defaultTar, buf[:n])
			if size-offsize <= defaultReadFileBufferSize {
				n, err = source.Read(buf[:size-offsize])
				if err != nil && err != io.EOF {
					return err
				}
				s.node.storage.AppendWrite(defaultTar, buf[:n])
				break
			}
		}
	} else {
		buf, err = s.node.storage.Load(name)
		if err != nil {
			return err
		}
		s.node.storage.AppendWrite(defaultTar, buf)
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
	if size > defaultReadFileBufferSize {
		buf = make([]byte, defaultReadFileBufferSize)
		for {
			n, err = source.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			s.node.storage.SeekWrite(name, offsize, buf[:n])
			offsize += defaultReadFileBufferSize
			if size-offsize <= defaultReadFileBufferSize {
				n, err = source.Read(buf[:size-offsize])
				if err != nil && err != io.EOF {
					return err
				}
				s.node.storage.SeekWrite(name, offsize, buf[:n])
				break
			}
		}
	} else {
		buf = make([]byte, size)
		_, err = source.Read(buf)
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
	if !s.node.storage.Exists(defaultLastIncludedIndex) {
		s.lastIncludedIndex.save()
		return errors.New(defaultLastIncludedIndex + " file is not existed")
	}
	if !s.node.storage.Exists(defaultLastIncludedTerm) {
		s.lastIncludedTerm.save()
		return errors.New(defaultLastIncludedTerm + " file is not existed")
	}
	if !s.node.storage.Exists(defaultSnapshot) {
		return errors.New(defaultSnapshot + " file is not existed")
	}
	//if !s.node.storage.Exists(DefaultIndex) {
	//	return errors.New(DefaultIndex + " file is not existed")
	//}
	//if !s.node.storage.Exists(DefaultLog) {
	//	return errors.New(DefaultLog + " file is not existed")
	//}
	s.done = false
	lastTarIndex := s.lastTarIndex.ID()
	s.lastTarIndex.Set(s.lastIncludedIndex.ID())
	s.node.storage.Truncate(defaultTar, 0)
	s.AppendFile(defaultLastTarIndex)
	s.AppendFile(defaultLastIncludedIndex)
	s.AppendFile(defaultLastIncludedTerm)
	s.AppendFile(defaultSnapshot)
	//s.AppendFile(DefaultIndex)
	//s.AppendFile(DefaultLog)
	if s.gzip {
		s.gz()
	}
	logger.Tracef("snapshotReadWriter.tar %s lastTarIndex %d==>%d", s.node.address, lastTarIndex, s.lastTarIndex.ID())
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
	//logger.Tracef("snapshotReadWriter.untar gzip %t", s.gzip)
	if s.gzip {
		s.ungz()
	}
	source, err := s.node.storage.FileReader(defaultTar)
	if err != nil {
		return err
	}
	defer source.Close()
	s.RecoverFile(source, defaultLastTarIndex)
	s.RecoverFile(source, defaultLastIncludedIndex)
	s.RecoverFile(source, defaultLastIncludedTerm)
	s.RecoverFile(source, s.name)
	//s.RecoverFile(source, DefaultIndex)
	//s.RecoverFile(source, DefaultLog)
	if !s.node.isLeader() {
		if s.node.storage.Exists(defaultTar) {
			s.node.storage.Rm(defaultTar)
		}
		if s.node.storage.Exists(defaultLastTarIndex) {
			s.node.storage.Rm(defaultLastTarIndex)
		}
	}
	if s.gzip {
		if s.node.storage.Exists(defaultTarGz) {
			s.node.storage.Rm(defaultTarGz)
		}
	}
	return nil
}
func (s *snapshotReadWriter) gz() error {
	dest, err := s.node.storage.FileWriter(defaultTarGz)
	if err != nil {
		return err
	}
	defer dest.Close()
	source, err := s.node.storage.FileReader(defaultTar)
	if err != nil {
		return err
	}
	defer source.Close()
	size, err := s.node.storage.Size(defaultTar)
	if err != nil {
		return err
	}
	writer := gzip.NewWriter(dest)
	defer writer.Close()
	buf := make([]byte, defaultReadFileBufferSize)
	var offsize int64 = 0
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		writer.Write(buf[:n])
		offsize += defaultReadFileBufferSize
		if size-offsize <= defaultReadFileBufferSize {
			n, err = source.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			writer.Write(buf[:n])
			break
		}
	}
	writer.Flush()
	return nil
}
func (s *snapshotReadWriter) ungz() error {
	dest, err := s.node.storage.FileWriter(defaultTar)
	if err != nil {
		return err
	}
	defer dest.Close()
	source, err := s.node.storage.FileReader(defaultTarGz)
	if err != nil {
		return err
	}
	defer source.Close()
	reader, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	defer reader.Close()
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
		if s.node.install() && s.lastIncludedIndex.ID() > s.lastTarIndex.ID() && s.node.isLeader() {
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
