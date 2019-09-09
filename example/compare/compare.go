package main

import (
	"os"
	"bytes"
	"io"
	"fmt"
)

func main() {
	f:=comparefile("/Users/huangmeng/go/src/defalut.raft.2/log","/Users/huangmeng/go/src/defalut.raft.3/log")
	fmt.Println(f)
}

func compare(spath, dpath string) bool {
	sinfo, err := os.Lstat(spath)
	if err != nil {
		return false
	}
	dinfo, err := os.Lstat(dpath)
	if err != nil {
		return false
	}
	if sinfo.Size() != dinfo.Size() || !sinfo.ModTime().Equal(dinfo.ModTime()) {
		return false
	}
	return comparefile(spath, dpath)
}

func comparefile(spath, dpath string) bool {
	sFile, err := os.Open(spath)
	if err != nil {
		return false
	}
	dFile, err := os.Open(dpath)
	if err != nil {
		return false
	}
	b := comparebyte(sFile, dFile)
	sFile.Close()
	dFile.Close()
	return b
}
func comparebyte(sfile *os.File, dfile *os.File) bool {
	var sbyte []byte = make([]byte, 1024)
	var dbyte []byte = make([]byte, 1024)
	var serr, derr error
	for {
		_, serr = sfile.Read(sbyte)
		_, derr = dfile.Read(dbyte)
		if serr != nil || derr != nil {
			if serr != derr {
				return false
			}
			if serr == io.EOF {
				break
			}
		}
		if bytes.Equal(sbyte,dbyte) {
			continue
		}
		fmt.Printf("%s\n%s\n",string(sbyte),string(dbyte))
		return false
	}
	return true
}
