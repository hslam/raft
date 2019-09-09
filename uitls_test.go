package raft

import "testing"

func TestUint64Encode(t *testing.T)  {
	var Offset uint64=104
	offsetBytes := uint64ToBytes(Offset)
	offset:= bytesToUint64(offsetBytes)
	if Offset!=offset{
		t.Errorf("%d %d",Offset,offset)
	}
}
