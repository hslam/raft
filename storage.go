// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"syscall"
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

func (s *storage) Sync(fileName string) error {
	filePath := path.Join(s.dataDir, fileName)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
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

func (s *storage) SafeExists(fileName string) bool {
	filePath := path.Join(s.dataDir, fileName)
	var exist = true
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		exist = s.SafeRecover(fileName)
	}
	return exist
}

func (s *storage) SafeRecover(fileName string) bool {
	filePath := path.Join(s.dataDir, fileName)
	tmpFilePath := path.Join(s.dataDir, fileName+defaultTmp)
	if _, err := os.Stat(tmpFilePath); os.IsNotExist(err) {
		return false
	}
	err := os.Rename(tmpFilePath, filePath)
	if err != nil {
		return false
	}
	return true
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

func (s *storage) RmDir(dirName string) {
	if err := os.RemoveAll(dirName); err != nil {
		logger.Errorf("Unable to RmDir path: %v", err)
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

func (s *storage) MD5(fileName string) string {
	filePath := path.Join(s.dataDir, fileName)
	file, err := os.Open(filePath)
	defer file.Close()
	if err == nil {
		hash := md5.New()
		io.Copy(hash, file)
		return fmt.Sprintf("%x", hash.Sum(nil))
	}
	return ""
}

func (s *storage) MD5Bytes(fileName string) []byte {
	filePath := path.Join(s.dataDir, fileName)
	file, err := os.Open(filePath)
	defer file.Close()
	if err == nil {
		hash := md5.New()
		io.Copy(hash, file)
		return hash.Sum(nil)
	}
	return []byte{}
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

func (s *storage) AppendWrite(fileName string, data []byte) error {
	filePath := path.Join(s.dataDir, fileName)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
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

func (s *storage) FileReader(fileName string) (*os.File, error) {
	filePath := path.Join(s.dataDir, fileName)
	return os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, 0600)
}

func (s *storage) FileWriter(fileName string) (*os.File, error) {
	filePath := path.Join(s.dataDir, fileName)
	return os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0600)
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

func (s *storage) TruncateTop(fileName string, size uint64) error {
	filePath := path.Join(s.dataDir, fileName)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	fileInfo, err := f.Stat()
	if err != nil {
		return err
	}
	mmap, err := syscall.Mmap(int(f.Fd()), 0, int(fileInfo.Size()), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	copy(mmap[0:fileInfo.Size()-int64(size)], mmap[size:])
	err = syscall.Munmap(mmap)
	if err != nil {
		return err
	}
	err = f.Truncate(fileInfo.Size() - int64(size))
	if err != nil {
		return err
	}
	return f.Sync()
}

func (s *storage) Copy(srcName string, dstName string, offset, size uint64) error {
	srcPath := path.Join(s.dataDir, srcName)
	dstPath := path.Join(s.dataDir, dstName)
	srcFile, err := os.OpenFile(srcPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	err = dstFile.Truncate(int64(size))
	if err != nil {
		return err
	}
	dstMmap, err := syscall.Mmap(int(dstFile.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	srcMmap, err := syscall.Mmap(int(srcFile.Fd()), 0, int(offset+size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	n := copy(dstMmap, srcMmap[offset:])
	if n != int(size) {
		return errors.New("error length")
	}
	err = syscall.Munmap(srcMmap)
	if err != nil {
		return err
	}
	err = syscall.Munmap(dstMmap)
	if err != nil {
		return err
	}
	return dstFile.Sync()
}

func (s *storage) AppendCopy(srcName string, dstName string, size uint64) error {
	srcPath := path.Join(s.dataDir, srcName)
	dstPath := path.Join(s.dataDir, dstName)
	srcFile, err := os.OpenFile(srcPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	dstFileInfo, err := dstFile.Stat()
	if err != nil {
		return err
	}
	dstSize := dstFileInfo.Size()
	err = dstFile.Truncate(dstSize + int64(size))
	if err != nil {
		return err
	}
	dstMmap, err := syscall.Mmap(int(dstFile.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	srcSize := srcFileInfo.Size()
	srcMmap, err := syscall.Mmap(int(srcFile.Fd()), 0, int(srcSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	copy(dstMmap[dstSize:], srcMmap[:size])
	err = syscall.Munmap(srcMmap)
	if err != nil {
		return err
	}
	err = syscall.Munmap(dstMmap)
	if err != nil {
		return err
	}
	return dstFile.Sync()
}

func (s *storage) Backup(srcName string, dstName string, size uint64) error {
	srcPath := path.Join(s.dataDir, srcName)
	dstPath := path.Join(s.dataDir, dstName)
	srcFile, err := os.OpenFile(srcPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	err = dstFile.Truncate(int64(size))
	if err != nil {
		return err
	}
	dstMmap, err := syscall.Mmap(int(dstFile.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	srcSize := srcFileInfo.Size()
	srcMmap, err := syscall.Mmap(int(srcFile.Fd()), 0, int(srcSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	copy(dstMmap[0:], srcMmap[:size])
	copy(srcMmap[0:], srcMmap[size:])
	err = syscall.Munmap(srcMmap)
	if err != nil {
		return err
	}
	err = syscall.Munmap(dstMmap)
	if err != nil {
		return err
	}
	err = srcFile.Truncate(srcSize - int64(size))
	if err != nil {
		return err
	}
	err = srcFile.Sync()
	if err != nil {
		return err
	}
	return dstFile.Sync()
}

func (s *storage) AppendBackup(srcName string, dstName string, size uint64) error {
	srcPath := path.Join(s.dataDir, srcName)
	dstPath := path.Join(s.dataDir, dstName)
	srcFile, err := os.OpenFile(srcPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	dstFileInfo, err := dstFile.Stat()
	if err != nil {
		return err
	}
	dstSize := dstFileInfo.Size()
	err = dstFile.Truncate(dstSize + int64(size))
	if err != nil {
		return err
	}
	dstMmap, err := syscall.Mmap(int(dstFile.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	srcSize := srcFileInfo.Size()
	srcMmap, err := syscall.Mmap(int(srcFile.Fd()), 0, int(srcSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	copy(dstMmap[dstSize:], srcMmap[:size])
	copy(srcMmap[0:], srcMmap[size:])
	err = syscall.Munmap(srcMmap)
	if err != nil {
		return err
	}
	err = syscall.Munmap(dstMmap)
	if err != nil {
		return err
	}
	err = srcFile.Truncate(srcSize - int64(size))
	if err != nil {
		return err
	}
	err = srcFile.Sync()
	if err != nil {
		return err
	}
	return dstFile.Sync()
}
