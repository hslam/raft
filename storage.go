package raft

import (
	"io"
	"os"
)
func checkIsExist(name string) bool {
	var exist = true
	if _, err := os.Stat(name); os.IsNotExist(err) {
		exist = false
	}
	return exist
}
func mkdir(dir_name string) {
	if err := os.MkdirAll(dir_name, 0744); err != nil {
		Errorf("Unable to create path: %v", err)
	}
}

func writeToFile(file_name string, data []byte ) error {
	f , err := os.OpenFile(file_name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
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

