package raft

import (
	"testing"
)

func TestMeta(t *testing.T) {
	meta := &Meta{}
	meta.Index = 1
	meta.Term = 1
	meta.Position = 10
	meta.Offset = 10
	{
		buf := make([]byte, 32)
		b, err := meta.Marshal(buf)
		if err != nil {
			t.Error(err)
		}
		meta1 := &Meta{}
		meta1.Unmarshal(b)
		if meta1.Index != meta.Index || meta1.Term != meta.Term || meta1.Position != meta.Position || meta1.Offset != meta.Offset {
			t.Error("error")
		}
		if len(b) != 32 {
			t.Errorf("%d !=32 ", len(b))
		}
	}
	{
		b, err := meta.Marshal(nil)
		if err != nil {
			t.Error(err)
		}
		meta1 := &Meta{}
		meta1.Unmarshal(b)
		if meta1.Index != meta.Index || meta1.Term != meta.Term || meta1.Position != meta.Position || meta1.Offset != meta.Offset {
			t.Error("error")
		}
		if len(b) != 32 {
			t.Errorf("%d !=32 ", len(b))
		}
	}
}

//BenchmarkMeta 128656268	         9.37 ns/op
func BenchmarkMeta(t *testing.B) {
	meta := &Meta{}
	meta.Index = 1
	meta.Term = 1
	meta.Position = 10
	meta.Offset = 10
	buf := make([]byte, 32)
	var data []byte
	meta1 := &Meta{}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		data, _ = meta.Marshal(buf)
		meta1.Unmarshal(data)
	}
}

//BenchmarkMetaNoBuffer 35.4 ns/op
func BenchmarkMetaNoBuffer(t *testing.B) {
	meta := &Meta{}
	meta.Index = 1
	meta.Term = 1
	meta.Position = 10
	meta.Offset = 10
	var data []byte
	meta1 := &Meta{}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		data, _ = meta.Marshal(nil)
		meta1.Unmarshal(data)
	}
}
