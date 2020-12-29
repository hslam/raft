// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type storage struct {
	dataDir string
}

var errSeeker = errors.New("seeker can't seek")

func newStorage(dataDir string) *storage {
	s := &storage{
		dataDir: dataDir,
	}
	s.MkDir(s.dataDir)
	return s
}

func (s *storage) FilePath(fileName string) string {
	return path.Join(s.dataDir, fileName)
}

func (s *storage) IsEmpty(fileName string) bool {
	var empty bool
	if num, err := s.Size(fileName); err != nil || num == 0 || !s.Exists(fileName) {
		empty = true
	} else {
		empty = false
	}
	return empty
}

func (s *storage) Exists(fileName string) bool {
	filePath := path.Join(s.dataDir, fileName)
	var exist = true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func (s *storage) MkDir(dirName string) {
	if err := os.MkdirAll(dirName, 0744); err != nil {
		logger.Errorf("Unable to create path: %v", err)
	}
}

func (s *storage) Rm(fileName string) {
	filePath := path.Join(s.dataDir, fileName)
	if err := os.Remove(filePath); err != nil {
		logger.Errorf("Unable to Rm path: %v", err)
	}
}

func (s *storage) SafeOverWrite(fileName string, data []byte) error {
	filePath := path.Join(s.dataDir, fileName)
	tmpFilePath := path.Join(s.dataDir, fileName+defaultTmp)
	flushFilePath := path.Join(s.dataDir, fileName+defaultFlush)
	defer func() {
		os.Rename(filePath, tmpFilePath)
		os.Rename(flushFilePath, filePath)
		os.Remove(tmpFilePath)
	}()
	f, err := os.OpenFile(flushFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := f.Write(data)
	if err != nil {
		return err
	} else if n < len(data) {
		return io.ErrShortWrite
	}
	return f.Sync()
}

func (s *storage) Rename(oldName, newName string) error {
	oldPath := path.Join(s.dataDir, oldName)
	newPath := path.Join(s.dataDir, newName)
	return os.Rename(oldPath, newPath)
}

func (s *storage) Load(fileName string) ([]byte, error) {
	filePath := path.Join(s.dataDir, fileName)
	return ioutil.ReadFile(filePath)
}

func (s *storage) Size(fileName string) (int64, error) {
	filePath := path.Join(s.dataDir, fileName)
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Seek(0, os.SEEK_END)
}

func (s *storage) OverWrite(fileName string, data []byte) error {
	filePath := path.Join(s.dataDir, fileName)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := f.Write(data)
	if err != nil {
		return err
	} else if n < len(data) {
		return io.ErrShortWrite
	}
	return f.Sync()
}

func (s *storage) SeekWrite(fileName string, cursor uint64, data []byte) error {
	filePath := path.Join(s.dataDir, fileName)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	ret, _ := f.Seek(int64(cursor), os.SEEK_SET)
	n, err := f.WriteAt(data, ret)
	if err != nil {
		return err
	} else if n < len(data) {
		return io.ErrShortWrite
	}
	return f.Sync()
}

func (s *storage) SeekWriteNoSync(fileName string, cursor uint64, data []byte) error {
	filePath := path.Join(s.dataDir, fileName)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	ret, _ := f.Seek(int64(cursor), os.SEEK_SET)
	n, err := f.WriteAt(data, ret)
	if err != nil {
		return err
	} else if n < len(data) {
		return io.ErrShortWrite
	}
	return nil
}

func (s *storage) SeekRead(fileName string, cursor uint64, b []byte) (n int, err error) {
	filePath := path.Join(s.dataDir, fileName)
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	ret, _ := f.Seek(int64(cursor), os.SEEK_SET)
	return f.ReadAt(b, ret)
}

func (s *storage) Truncate(fileName string, size uint64) error {
	filePath := path.Join(s.dataDir, fileName)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	err = f.Truncate(int64(size))
	if err != nil {
		return err
	}
	return f.Sync()
}
