package raft


import (
	"testing"
)


func BenchmarkIndexCodecEncode(t *testing.B) {
	meta:=&Meta{}
	meta.Index=1
	meta.Term=1
	meta.Ret=10
	meta.Offset=10
	metaStorage:=&MetaStorage{Metas:make([]*Meta,0)}
	codec:=&ProtoCodec{}
	for i:=0;i<1000;i++{
		metaStorage.Metas=append(metaStorage.Metas, meta)
	}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		codec.Encode(metaStorage)
	}
}
func BenchmarkIndexCodecDecode(t *testing.B) {
	meta:=&Meta{}
	meta.Index=1
	meta.Term=1
	meta.Ret=10
	meta.Offset=10
	metaStorage:=&MetaStorage{Metas:make([]*Meta,0)}
	codec:=&ProtoCodec{}
	for i:=0;i<1000;i++{
		metaStorage.Metas=append(metaStorage.Metas, meta)
	}
	b,_:=codec.Encode(metaStorage)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		codec.Decode(b,metaStorage)
	}
}
func BenchmarkIndexCodecEncodeDecode(t *testing.B) {
	meta:=&Meta{}
	meta.Index=1
	meta.Term=1
	meta.Ret=10
	meta.Offset=10
	metaStorage:=&MetaStorage{Metas:make([]*Meta,0)}
	codec:=&ProtoCodec{}
	for i:=0;i<1000;i++{
		metaStorage.Metas=append(metaStorage.Metas, meta)
	}
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		b,_:=codec.Encode(metaStorage)
		codec.Decode(b,metaStorage)
	}
}

//go test -v -bench=.  -benchmem -benchtime=1s
//=== RUN   TestMetaEncode
//--- PASS: TestMetaEncode (0.00s)
//goos: darwin
//goarch: amd64
//pkg: hslam.com/git/x/raft
 //BenchmarkIndexEncode-4               	   10000	    120250 ns/op	  209568 B/op	    3019 allocs/op
//BenchmarkIndexDecode-4               	   50000	     24503 ns/op	       0 B/op	       0 allocs/op
//BenchmarkIndexEncodeDecode-4         	   10000	    144627 ns/op	  209568 B/op	    3019 allocs/op
//BenchmarkIndexCodecEncode-4          	   20000	     99456 ns/op	   10240 B/op	       1 allocs/op
//BenchmarkIndexCodecDecode-4          	   10000	    124158 ns/op	   80376 B/op	    1011 allocs/op
//BenchmarkIndexCodecEncodeDecode-4    	   10000	    225282 ns/op	   90616 B/op	    1012 allocs/op