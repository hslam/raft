// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"errors"
	"github.com/hslam/atomic"
	"github.com/hslam/tar"
	"io"
	"sync"
	"time"
)

// Snapshot saves a snapshot and recovers from a snapshot.
type Snapshot interface {
	// Save writes snapshot data to w until there's no more data to write or
	// when an error occurs. The return value n is the number of bytes
	// written. Any error encountered during the write is also returned.
	Save(w io.Writer) (n int, err error)
	// Recover reads snapshot data from r until EOF or error.
	// The return value n is the number of bytes read.
	// Any error except io.EOF encountered during the read is also returned.
	Recover(r io.Reader) (n int, err error)
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
	finish            bool
	gzip              bool
	done              chan struct{}
	closed            int32
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
		done:              make(chan struct{}, 1),
	}
	go s.run()
	return s
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
	if !s.node.storage.Exists(defaultConfig) {
		return errors.New(defaultConfig + " file is not existed")
	}
	if !s.node.storage.Exists(defaultSnapshot) {
		return errors.New(defaultSnapshot + " file is not existed")
	}
	s.finish = false
	lastTarIndex := s.lastTarIndex.ID()
	s.lastTarIndex.Set(s.lastIncludedIndex.ID())
	if s.gzip {
		tar.Targz(s.node.storage.FilePath(defaultTarGz),
			s.node.storage.FilePath(defaultLastTarIndex),
			s.node.storage.FilePath(defaultConfig),
			s.node.storage.FilePath(defaultSnapshot),
		)
	} else {
		tar.Tar(s.node.storage.FilePath(defaultTar),
			s.node.storage.FilePath(defaultLastTarIndex),
			s.node.storage.FilePath(defaultConfig),
			s.node.storage.FilePath(defaultSnapshot),
		)
	}
	s.finish = true
	s.node.logger.Tracef("snapshotReadWriter.tar %s lastTarIndex %d==>%d", s.node.address, lastTarIndex, s.lastTarIndex.ID())
	return nil
}

func (s *snapshotReadWriter) untar() error {
	if s.node.storage.IsEmpty(s.FileName()) {
		return errors.New(s.FileName() + " file is empty")
	}
	if !s.finish {
		return nil
	}
	s.node.logger.Tracef("snapshotReadWriter.untar gzip %t dir %s", s.gzip, s.node.storage.dataDir)
	if s.gzip {
		tar.Untargz(s.node.storage.FilePath(defaultTarGz), s.node.storage.dataDir)
		s.node.storage.Rm(defaultTarGz)
	} else {
		tar.Untar(s.node.storage.FilePath(defaultTar), s.node.storage.dataDir)
		s.node.storage.Rm(defaultTar)
	}
	return nil
}

func (s *snapshotReadWriter) clear() error {
	return s.node.storage.Truncate(s.FileName(), 0)
}

func (s *snapshotReadWriter) canTar() bool {
	s.mut.RLock()
	tarWork := s.tarWork
	s.mut.RUnlock()
	return tarWork
}

func (s *snapshotReadWriter) disableTar() {
	s.mut.Lock()
	s.tarWork = false
	s.mut.Unlock()
}

func (s *snapshotReadWriter) enableTar() {
	s.mut.Lock()
	s.tarWork = true
	s.mut.Unlock()
}

func (s *snapshotReadWriter) run() {
	for {
		select {
		case <-s.ticker.C:
			if s.node.install() && s.lastIncludedIndex.ID() > s.lastTarIndex.ID() && s.node.isLeader() {
				s.Tar()
			}
		case <-s.done:
			return
		}
	}
}

func (s *snapshotReadWriter) Stop() {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		s.ticker.Stop()
		close(s.done)
	}
}
