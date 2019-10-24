package raft

import (
	"io"
	"os"
	"path"
	"io/ioutil"
	"crypto/md5"
	"fmt"
	"errors"
	"syscall"
)

type Storage struct {
	node							*Node
	data_dir    					string
}

var errSeeker = errors.New("seeker can't seek")

func newStorage(node *Node,data_dir string)*Storage {
	s:=&Storage{
		node:node,
		data_dir:data_dir,
	}
	s.MkDir(s.data_dir)
	return s
}
func (s *Storage)Exists(file_name string) bool {
	file_path := path.Join(s.data_dir, file_name)
	var exist = true
	if _, err := os.Stat(file_path); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
func (s *Storage)IsEmpty(file_name string) bool {
	var empty bool
	if num,err:=s.Size(file_name);err!=nil||num==0||!s.Exists(file_name){
		empty = true
	}else {
		empty = false
	}
	return empty
}
func (s *Storage)SafeExists(file_name string) bool {
	file_path := path.Join(s.data_dir, file_name)
	var exist = true
	if _, err := os.Stat(file_path); os.IsNotExist(err) {
		exist = s.SafeRecover(file_name)
	}
	return exist
}
func (s *Storage)SafeRecover(file_name string) bool {
	file_path := path.Join(s.data_dir, file_name)
	tmp_file_path := path.Join(s.data_dir, file_name+DefaultTmp)
	if _, err := os.Stat(tmp_file_path); os.IsNotExist(err) {
		return false
	}
	err:=os.Rename(tmp_file_path, file_path)
	if err!=nil{
		return false
	}
	return true
}
func (s *Storage)MkDir(dir_name string) {
	if err := os.MkdirAll(dir_name, 0744); err != nil {
		Errorf("Unable to create path: %v", err)
	}
}
func (s *Storage)RmDir(dir_name string) {
	if err := os.RemoveAll(dir_name); err != nil {
		Errorf("Unable to RmDir path: %v", err)
	}
}
func (s *Storage)Rm(file_name string) {
	file_path := path.Join(s.data_dir, file_name)
	if err := os.Remove(file_path); err != nil {
		Errorf("Unable to Rm path: %v", err)
	}
}
func (s *Storage)SafeOverWrite(file_name string, data []byte ) error {
	file_path := path.Join(s.data_dir, file_name)
	if s.node.fast(){
		f , err := os.OpenFile(file_path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		n, err := f.Write(data)
		if err != nil {
			return err
		}else if n < len(data) {
			return io.ErrShortWrite
		}
		return nil
	}else {
		tmp_file_path := path.Join(s.data_dir, file_name+DefaultTmp)
		defer func() {
			os.Rename(tmp_file_path, file_path)
		}()
		f , err := os.OpenFile(tmp_file_path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		n, err := f.Write(data)
		if err != nil {
			return err
		}else if n < len(data) {
			return io.ErrShortWrite
		}
		return f.Sync()
	}
}
func (s *Storage) Rename(old_name, new_name string) error {
	old_path := path.Join(s.data_dir, old_name)
	new_path := path.Join(s.data_dir, new_name)
	return os.Rename(old_path, new_path)
}
func (s *Storage) Load(file_name string) ([]byte,error) {
	file_path := path.Join(s.data_dir, file_name)
	return ioutil.ReadFile(file_path)
}

func (s *Storage) Size(file_name string) (int64,error) {
	file_path := path.Join(s.data_dir, file_name)
	f , err := os.OpenFile(file_path, os.O_RDONLY, 0600)
	if err != nil {
		return 0,err
	}
	defer f.Close()
	return f.Seek(0, os.SEEK_END)
}

func (s *Storage) MD5(file_name string) string{
	file_path := path.Join(s.data_dir, file_name)
	file, err := os.Open(file_path)
	defer file.Close()
	if err == nil {
		hash := md5.New()
		io.Copy(hash, file)
		return fmt.Sprintf("%x", hash.Sum(nil))
	}
	return ""
}
func (s *Storage) MD5Bytes(file_name string) []byte{
	file_path := path.Join(s.data_dir, file_name)
	file, err := os.Open(file_path)
	defer file.Close()
	if err == nil {
		hash := md5.New()
		io.Copy(hash, file)
		return hash.Sum(nil)
	}
	return []byte{}
}
func (s *Storage)OverWrite(file_name string, data []byte ) error {
	file_path := path.Join(s.data_dir, file_name)
	f , err := os.OpenFile(file_path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := f.Write(data)
	if err != nil {
		return err
	}else if n < len(data) {
		return io.ErrShortWrite
	}
	if s.node.fast(){
		return nil
	}
	return f.Sync()
}

func (s *Storage)AppendWrite(file_name string, data []byte) error {
	file_path := path.Join(s.data_dir, file_name)
	f, err := os.OpenFile(file_path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := f.Write(data)
	if err != nil {
		return err
	}else if n < len(data) {
		return io.ErrShortWrite
	}
	if s.node.fast(){
		return nil
	}
	return f.Sync()
}
func (s *Storage)SeekWrite(file_name string,cursor uint64,data []byte) error {
	file_path := path.Join(s.data_dir, file_name)
	f, err := os.OpenFile(file_path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	ret, _ := f.Seek(int64(cursor), os.SEEK_SET)
	n, err := f.WriteAt(data, ret)
	if err != nil {
		return err
	}else if n < len(data) {
		return io.ErrShortWrite
	}
	if s.node.fast(){
		return nil
	}
	return f.Sync()
}

func (s *Storage) SeekRead(file_name string,cursor uint64,b []byte)(n int, err error) {
	file_path := path.Join(s.data_dir, file_name)
	f , err := os.OpenFile(file_path, os.O_RDONLY, 0600)
	if err != nil {
		return 0,err
	}
	defer f.Close()
	ret, _ := f.Seek(int64(cursor), os.SEEK_SET)
	return f.ReadAt(b, ret)
}

func (s *Storage)FileReader(file_name string) (*os.File,error) {
	file_path := path.Join(s.data_dir, file_name)
	return os.OpenFile(file_path, os.O_CREATE|os.O_RDONLY, 0600)
}

func (s *Storage)FileWriter(file_name string) (*os.File,error) {
	file_path := path.Join(s.data_dir, file_name)
	return os.OpenFile(file_path, os.O_CREATE|os.O_WRONLY, 0600)
}
func (s *Storage)Truncate(file_name string,size uint64) error {
	file_path := path.Join(s.data_dir, file_name)
	f, err := os.OpenFile(file_path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	err= f.Truncate(int64(size))
	if err != nil {
		return err
	}
	return f.Sync()
}
func (s *Storage)TruncateTop(file_name string,size uint64) error {
	file_path := path.Join(s.data_dir, file_name)
	f , err := os.OpenFile(file_path, os.O_CREATE|os.O_RDWR, 0600)
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
	copy(mmap[0:], mmap[size:])
	err = syscall.Munmap(mmap)
	if err != nil {
		return err
	}
	err=f.Truncate(fileInfo.Size() - int64(size))
	if err != nil {
		return err
	}
	return f.Sync()
}
func (s *Storage)Copy(src_name string,dst_name string,size uint64) error {
	src_path := path.Join(s.data_dir, src_name)
	dst_path := path.Join(s.data_dir, dst_name)
	src_file , err := os.OpenFile(src_path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer src_file.Close()
	dst_file , err := os.OpenFile(dst_path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer dst_file.Close()
	err=dst_file.Truncate(int64(size))
	if err != nil {
		return err
	}
	dst_mmap, err := syscall.Mmap(int(dst_file.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	src_fileInfo, err := src_file.Stat()
	if err != nil {
		return err
	}
	src_size:=src_fileInfo.Size()
	src_mmap, err := syscall.Mmap(int(src_file.Fd()), 0, int(src_size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	copy(dst_mmap[0:], src_mmap[:size])
	err = syscall.Munmap(src_mmap)
	if err != nil {
		return err
	}
	err = syscall.Munmap(dst_mmap)
	if err != nil {
		return err
	}
	return dst_file.Sync()
}
func (s *Storage)AppendCopy(src_name string,dst_name string,size uint64) error {
	src_path := path.Join(s.data_dir, src_name)
	dst_path := path.Join(s.data_dir, dst_name)
	src_file , err := os.OpenFile(src_path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer src_file.Close()
	dst_file , err := os.OpenFile(dst_path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer dst_file.Close()
	dst_fileInfo, err := dst_file.Stat()
	if err != nil {
		return err
	}
	dst_size:=dst_fileInfo.Size()
	err=dst_file.Truncate(dst_size+int64(size))
	if err != nil {
		return err
	}
	dst_mmap, err := syscall.Mmap(int(dst_file.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	src_fileInfo, err := src_file.Stat()
	if err != nil {
		return err
	}
	src_size:=src_fileInfo.Size()
	src_mmap, err := syscall.Mmap(int(src_file.Fd()), 0, int(src_size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	copy(dst_mmap[dst_size:], src_mmap[:size])
	err = syscall.Munmap(src_mmap)
	if err != nil {
		return err
	}
	err = syscall.Munmap(dst_mmap)
	if err != nil {
		return err
	}
	return dst_file.Sync()
}
func (s *Storage)Backup(src_name string,dst_name string,size uint64) error {
	src_path := path.Join(s.data_dir, src_name)
	dst_path := path.Join(s.data_dir, dst_name)
	src_file , err := os.OpenFile(src_path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer src_file.Close()
	dst_file , err := os.OpenFile(dst_path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer dst_file.Close()
	err=dst_file.Truncate(int64(size))
	if err != nil {
		return err
	}
	dst_mmap, err := syscall.Mmap(int(dst_file.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	src_fileInfo, err := src_file.Stat()
	if err != nil {
		return err
	}
	src_size:=src_fileInfo.Size()
	src_mmap, err := syscall.Mmap(int(src_file.Fd()), 0, int(src_size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	copy(dst_mmap[0:], src_mmap[:size])
	copy(src_mmap[0:], src_mmap[size:])
	err = syscall.Munmap(src_mmap)
	if err != nil {
		return err
	}
	err = syscall.Munmap(dst_mmap)
	if err != nil {
		return err
	}
	err=src_file.Truncate(src_size - int64(size))
	if err != nil {
		return err
	}
	err=src_file.Sync()
	if err != nil {
		return err
	}
	return dst_file.Sync()
}
func (s *Storage)AppendBackup(src_name string,dst_name string,size uint64) error {
	src_path := path.Join(s.data_dir, src_name)
	dst_path := path.Join(s.data_dir, dst_name)
	src_file , err := os.OpenFile(src_path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer src_file.Close()
	dst_file , err := os.OpenFile(dst_path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer dst_file.Close()
	dst_fileInfo, err := dst_file.Stat()
	if err != nil {
		return err
	}
	dst_size:=dst_fileInfo.Size()
	err=dst_file.Truncate(dst_size+int64(size))
	if err != nil {
		return err
	}
	dst_mmap, err := syscall.Mmap(int(dst_file.Fd()), 0, int(size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	src_fileInfo, err := src_file.Stat()
	if err != nil {
		return err
	}
	src_size:=src_fileInfo.Size()
	src_mmap, err := syscall.Mmap(int(src_file.Fd()), 0, int(src_size), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	copy(dst_mmap[dst_size:], src_mmap[:size])
	copy(src_mmap[0:], src_mmap[size:])
	err = syscall.Munmap(src_mmap)
	if err != nil {
		return err
	}
	err = syscall.Munmap(dst_mmap)
	if err != nil {
		return err
	}
	err=src_file.Truncate(src_size - int64(size))
	if err != nil {
		return err
	}
	err=src_file.Sync()
	if err != nil {
		return err
	}
	return dst_file.Sync()
}
