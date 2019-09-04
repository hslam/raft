package raft

import (
	"bytes"
)

type Index struct {
	index		uint64
	term		uint64
	ret			uint64
	offset		uint64
}

func (i Index)Encode()[]byte  {
	indexBytes := uint64ToBytes(i.index)
	termBytes :=uint64ToBytes(i.term)
	retBytes := uint64ToBytes(i.ret)
	offsetBytes := uint64ToBytes(i.offset)
	Bytes:=append(indexBytes,termBytes...)
	Bytes=append(Bytes,retBytes...)
	Bytes=append(termBytes,offsetBytes...)
	return Bytes
}

func (i Index)Decode(data []byte)  {
	var buf =bytes.NewBuffer(data)
	indexBytes:=make([]byte , 8)
	buf.Read(indexBytes)
	i.index=bytesToUint64(indexBytes)
	termBytes:=make([]byte , 8)
	buf.Read(termBytes)
	i.term= bytesToUint64(termBytes)
	retBytes:=make([]byte , 8)
	buf.Read(retBytes)
	i.ret= bytesToUint64(retBytes)
	offsetBytes:=make([]byte , 8)
	buf.Read(offsetBytes)
	i.offset= bytesToUint64(offsetBytes)
}
