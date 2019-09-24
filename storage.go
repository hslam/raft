package raft

import (
	"io"
	"os"
	"path"
	"io/ioutil"
	"crypto/md5"
	"fmt"
	"errors"
)

type Storage struct {
	data_dir    					string
}

var errSeeker = errors.New("seeker can't seek")

func newStorage(data_dir string)*Storage {
	s:=&Storage{
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
func (s *Storage)SafeOverWrite(file_name string, data []byte ) error {
	file_path := path.Join(s.data_dir, file_name)
	tmp_file_path := path.Join(s.data_dir, file_name+DefaultTmp)
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
	f.Sync()
	return os.Rename(tmp_file_path, file_path)
}
func (s *Storage) Load(file_name string) ([]byte,error) {
	file_path := path.Join(s.data_dir, file_name)
	return ioutil.ReadFile(file_path)
}

func (s *Storage) Size(file_name string) (int64,error) {
	file_path := path.Join(s.data_dir, file_name)
	f , err := os.OpenFile(file_path, os.O_RDONLY, 0600)
	if err != nil {
		return -1,err
	}
	defer f.Close()
	return f.Seek(0, os.SEEK_END)
}
func (s *Storage) MD5(file_name string) string{
	file_path := path.Join(s.data_dir, file_name)
	testFile := file_path
	file, inerr := os.Open(testFile)
	defer file.Close()
	if inerr == nil {
		md5h := md5.New()
		io.Copy(md5h, file)
		return fmt.Sprintf("%x", md5h.Sum(nil))
	}
	return ""
}
func (s *Storage) MD5Bytes(file_name string) []byte{
	file_path := path.Join(s.data_dir, file_name)
	testFile := file_path
	file, inerr := os.Open(testFile)
	defer file.Close()
	if inerr == nil {
		md5h := md5.New()
		io.Copy(md5h, file)
		return md5h.Sum(nil)
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
	return f.Sync()
}
func (s *Storage) SeekRead(file_name string,cursor uint64,b []byte)(err error) {
	file_path := path.Join(s.data_dir, file_name)
	f , err := os.OpenFile(file_path, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	ret, _ := f.Seek(int64(cursor), os.SEEK_SET)
	n, err := f.ReadAt(b, ret)
	if err != nil {
		return err
	}else if n < len(b) {
		return io.ErrShortBuffer
	}
	return nil
}