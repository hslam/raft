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
	s.Mkdir(s.data_dir)
	return s
}
func (s *Storage)CheckFile(name string) bool {
	var exist = true
	if _, err := os.Stat(name); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
func (s *Storage)Mkdir(dir_name string) {
	if err := os.MkdirAll(dir_name, 0744); err != nil {
		Errorf("Unable to create path: %v", err)
	}
}
func (s *Storage)Persistence(file_name string, data []byte ) error {
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
