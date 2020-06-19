package raft

import (
	"github.com/hslam/code"
)


func (meta *Meta)Encode()[]byte  {
	buf:=make([]byte,32)
	var offset uint64
	var n  uint64
	n = code.EncodeUint64(buf[offset:offset+8],meta.Index)
	offset+=n
	n = code.EncodeUint64(buf[offset:offset+8],meta.Term)
	offset+=n
	n = code.EncodeUint64(buf[offset:offset+8],meta.Position)
	offset+=n
	n = code.EncodeUint64(buf[offset:offset+8],meta.Offset)
	offset+=n
	return buf
}

func (meta *Meta)Decode(data []byte)  {
	var offset uint64
	var n uint64
	n=code.DecodeUint64(data[offset:offset+8],&meta.Index)
	offset+=n
	n=code.DecodeUint64(data[offset:offset+8],&meta.Term)
	offset+=n
	n=code.DecodeUint64(data[offset:offset+8],&meta.Position)
	offset+=n
	n=code.DecodeUint64(data[offset:offset+8],&meta.Offset)
	offset+=n
}

