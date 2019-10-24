package raft

import (
	"bytes"
)

const MetaSize  = 32

func (meta *Meta)Encode()[]byte  {
	indexBytes := uint64ToBytes(meta.Index)
	termBytes :=uint64ToBytes(meta.Term)
	retBytes := uint64ToBytes(meta.Position)
	offsetBytes := uint64ToBytes(meta.Offset)
	Bytes:=make([]byte,0,32)
	Bytes=append(Bytes,indexBytes...)
	Bytes=append(Bytes,termBytes...)
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
	meta.Position= bytesToUint64(retBytes)
	offsetBytes:=make([]byte , 8)
	buf.Read(offsetBytes)
	meta.Offset= bytesToUint64(offsetBytes)
}
