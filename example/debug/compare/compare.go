package main

import (
	"os"
	"bytes"
	"io"
	"fmt"
)

const MaxReadBuffer  = 1024*64
func main() {
	dir:="/Users/huangmeng/go/src/"
	nodes:=[]string{"1","2","3"}
	files:=[]string{"lastlogindex","index","log","commitindex","term","votefor"}
	prints:=[]string{"","\t","\t\t","","\t\t","\t"}
	if len(files)!=len(prints){
		return
	}
	for i:=1;i<len(nodes) ;i++  {
		for j:=0;j<len(files) ;j++  {
			path1:="defalut.raft."+nodes[0]+"/"+files[j]
			path2:="defalut.raft."+nodes[i]+"/"+files[j]
			equal,strs:=Compare(dir+path1,dir+path2,MaxReadBuffer)
			fmt.Printf("%s\t%s\t%s\t%s\t%t\n",path1,prints[j],path2,prints[j],equal)
			if !equal&&len(strs)==2{
				fmt.Printf("%s\n%s\n",strs[0],strs[1])
			}
		}
	}
}

func Compare(source_path, dest_path string,buffer_size int) (bool,[][]byte)  {
	if CompareInfo(source_path,dest_path){
		return CompareFile(source_path, dest_path,buffer_size)
	}
	return false,nil
}
func CompareInfo(source_path, dest_path string) (bool)  {
	source_info, err := os.Lstat(source_path)
	if err != nil {
		return false
	}
	dest_info, err := os.Lstat(dest_path)
	if err != nil {
		return false
	}
	if source_info.Size() != dest_info.Size() {
		return false
	}
	return true
}

func CompareFile(source_path, dest_path string,buffer_size int) (bool,[][]byte) {
	source_file, err := os.Open(source_path)
	if err != nil {
		return false,nil
	}
	defer source_file.Close()
	dest_file, err := os.Open(dest_path)
	if err != nil {
		return false,nil
	}
	defer dest_file.Close()
	b ,strs:= CompareBytes(source_file, dest_file,buffer_size)

	return b,strs
}
func CompareBytes(source_file *os.File, dest_file *os.File,buffer_size int) (bool,[][]byte) {
	var(
		source_buf = make([]byte, buffer_size)
		dest_buf = make([]byte, buffer_size)
		source_n, dest_n int
		source_err, dest_err error
	)
	for {
		source_n, source_err = source_file.Read(source_buf)
		dest_n, dest_err = dest_file.Read(dest_buf)
		if source_err != nil || dest_err != nil {
			if source_err != dest_err {
				return false,nil
			}
			if source_err == io.EOF {
				break
			}
		}
		if source_n != dest_n {
			return false,nil
		}
		if bytes.Equal(source_buf[:source_n],dest_buf[:dest_n]) {
			continue
		}
		return false,[][]byte{source_buf[:source_n],dest_buf[:dest_n]}
	}
	return true,nil
}
