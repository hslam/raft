package raft

import (
	"github.com/hslam/code"
)

type Meta struct {
	Index    uint64
	Term     uint64
	Position uint64
	Offset   uint64
}

func (meta *Meta) Marshal(buf []byte) ([]byte, error) {
	var size uint64 = 32
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		buf = make([]byte, size)
	}
	var offset uint64
	var n uint64
	n = code.EncodeUint64(buf[offset:offset+8], meta.Index)
	offset += n
	n = code.EncodeUint64(buf[offset:offset+8], meta.Term)
	offset += n
	n = code.EncodeUint64(buf[offset:offset+8], meta.Position)
	offset += n
	n = code.EncodeUint64(buf[offset:offset+8], meta.Offset)
	offset += n
	return buf[:offset], nil
}

func (meta *Meta) Unmarshal(data []byte) (uint64, error) {
	var offset uint64
	var n uint64
	n = code.DecodeUint64(data[offset:offset+8], &meta.Index)
	offset += n
	n = code.DecodeUint64(data[offset:offset+8], &meta.Term)
	offset += n
	n = code.DecodeUint64(data[offset:offset+8], &meta.Position)
	offset += n
	n = code.DecodeUint64(data[offset:offset+8], &meta.Offset)
	offset += n
	return offset, nil
}
