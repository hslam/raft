package raft

import (
	"io"
	"os"
	"path"
	"io/ioutil"
)

type Storage struct {
	data_dir    					string
}

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
	config_path := path.Join(s.data_dir, file_name)
	return ioutil.ReadFile(config_path)
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
	f.Sync()
	return err
}
func (s *Storage)SeekWrite(file_name string,offset int64,data []byte) error {
	file_path := path.Join(s.data_dir, file_name)
	f, err := os.OpenFile(file_path, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	ret, _ := f.Seek(offset, os.SEEK_SET)
	n, err := f.WriteAt(data, ret)
	if err != nil {
		return err
	}else if n < len(data) {
		return io.ErrShortWrite
	}
	f.Sync()
	return err
}
