package raft

import (
	"bytes"
)

func (meta *Meta)Encode()[]byte  {
	indexBytes := uint64ToBytes(meta.Index)
	termBytes :=uint64ToBytes(meta.Term)
	retBytes := uint64ToBytes(meta.Ret)
	offsetBytes := uint64ToBytes(meta.Offset)
	Bytes:=append(indexBytes,termBytes...)
	Bytes=append(Bytes,retBytes...)
	Bytes=append(Bytes,offsetBytes...)
	return Bytes
}

func (meta *Meta)Decode(data []byte)  {
	var buf =bytes.NewBuffer(data)
	indexBytes:=make([]byte , 8)
	buf.Read(indexBytes)
	meta.Index=bytesToUint64(indexBytes)
	termBytes:=make([]byte , 8)
	buf.Read(termBytes)
	meta.Term= bytesToUint64(termBytes)
	retBytes:=make([]byte , 8)
	buf.Read(retBytes)
	meta.Ret= bytesToUint64(retBytes)
	offsetBytes:=make([]byte , 8)
	buf.Read(offsetBytes)
	meta.Offset= bytesToUint64(offsetBytes)
}
