package raft

import (
	"testing"
)
func TestMetaEncode(t *testing.T) {
	meta:=&Meta{}
	meta.Index=1
	meta.Term=1
	meta.Ret=10
	meta.Offset=10
	b:=meta.Encode()
	meta1:=&Meta{}
	meta1.Decode(b)
	if meta1.Index!=meta.Index||meta1.Term!=meta.Term||meta1.Ret!=meta.Ret||meta1.Offset!=meta.Offset{
		t.Error(nil)
	}
	if len(b)!=32{
		t.Errorf("%d !=32 ",len(b))
	}
}
func TestMetaCodecEncode(t *testing.T) {
	meta:=&Meta{}
	meta.Index=1
	meta.Term=1
	meta.Ret=10
	meta.Offset=10
	codec:=&ProtoCodec{}
	b,_:=codec.Encode(meta)
	meta1:=&Meta{}
	codec.Decode(b,meta1)
	if meta1.Index!=meta.Index||meta1.Term!=meta.Term||meta1.Ret!=meta.Ret||meta1.Offset!=meta.Offset{
		t.Error(nil)
	}
	//var Max int64 = -1 ^ (-1 << 63)
	//meta.Index=uint64(Max)
	//meta.Term=uint64(Max)
	//meta.Ret=uint64(Max)
	//meta.Offset=uint64(Max)
	//b,_=codec.Encode(meta)
	//if len(b)>0{
	//	t.Error(len(b))
	//}
}

func BenchmarkMetaEncode(t *testing.B) {
	meta:=&Meta{}
	meta.Index=1
	meta.Term=1
	meta.Ret=10
	meta.Offset=10
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		meta.Encode()
	}
}
func BenchmarkMetaDecode(t *testing.B) {
	meta:=&Meta{}
	meta.Index=1
	meta.Term=1
	meta.Ret=10
	meta.Offset=10
	b:=meta.Encode()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		meta.Decode(b)
	}
}
func BenchmarkMetaEncodeDecode(t *testing.B) {
	meta:=&Meta{}
	meta.Index=1
	meta.Term=1
	meta.Ret=10
	meta.Offset=10
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		b:=meta.Encode()
		meta.Decode(b)
	}
}


func BenchmarkMetaCodecEncode(t *testing.B) {
	meta:=&Meta{}
	meta.Index=1
	meta.Term=1
	meta.Ret=10
	meta.Offset=10
	codec:=&ProtoCodec{}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		codec.Encode(meta)
	}
}
func BenchmarkMetaCodecDecode(t *testing.B) {
	meta:=&Meta{}
	meta.Index=1
	meta.Term=1
	meta.Ret=10
	meta.Offset=10
	codec:=&ProtoCodec{}
	b,_:=codec.Encode(meta)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		codec.Decode(b,meta)
	}
}
func BenchmarkMetaCodecEncodeDecode(t *testing.B) {
	meta:=&Meta{}
	meta.Index=1
	meta.Term=1
	meta.Ret=10
	meta.Offset=10
	codec:=&ProtoCodec{}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		b,_:=codec.Encode(meta)
		codec.Decode(b,meta)
	}
}

