package raft

import (
	"testing"
)

func TestStorageSafeOverWrite10B(t *testing.T) {
	b:=[]byte("123456789\n")
	storage:=newStorage(DefaultDataDir)
	err:=storage.SafeOverWrite("TestStorageSafeOverWrite10B",b)
	if err!=nil{
		t.Error(err)
	}
}

func BenchmarkStorageSafeOverWrite10B(t *testing.B) {
	b:=[]byte("123456789\n")
	storage:=newStorage(DefaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.SafeOverWrite("BenchmarkStorageSafeOverWrite10B",b)
	}
}

func BenchmarkStorageOverWrite10B(t *testing.B) {
	b:=[]byte("123456789\n")
	storage:=newStorage(DefaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.OverWrite("BenchmarkStorageOverWrite10B",b)
	}
}
func BenchmarkStorageAppendWrite10B(t *testing.B) {
	b:=[]byte("123456789\n")
	storage:=newStorage(DefaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.AppendWrite("BenchmarkStorageAppendWrite10B",b)
	}
}
func BenchmarkStorageSeekWrite10B(t *testing.B) {
	b:=[]byte("123456789\n")
	storage:=newStorage(DefaultDataDir)
	storage.AppendWrite("BenchmarkStorageSeekWrite10B",b)
	storage.AppendWrite("BenchmarkStorageSeekWrite10B",b)
	storage.AppendWrite("BenchmarkStorageSeekWrite10B",b)
	b=[]byte("000000000\n")

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.SeekWrite("BenchmarkStorageSeekWrite10B",10,b)
	}
}
func BenchmarkStorageAppendWrite1KB(t *testing.B) {
	b:=byte(48)
	bs:=[]byte{}
	for i:=0;i<1023;i++{
		bs=append(bs,b)
	}
	bs=append(bs,byte(10))
	storage:=newStorage(DefaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.AppendWrite("BenchmarkStorageAppendWrite1KB",bs)
	}
}
func BenchmarkStorageAppendWrite10KB(t *testing.B) {
	b:=byte(48)
	bs:=[]byte{}
	for i:=0;i<1024*10-1;i++{
		bs=append(bs,b)
	}
	bs=append(bs,byte(10))
	storage:=newStorage(DefaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.AppendWrite("BenchmarkStorageAppendWrite10KB",bs)
	}
}
func BenchmarkStorageAppendWrite100KB(t *testing.B) {
	b:=byte(48)
	bs:=[]byte{}
	for i:=0;i<1024*100-1;i++{
		bs=append(bs,b)
	}
	bs=append(bs,byte(10))
	storage:=newStorage(DefaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.AppendWrite("BenchmarkStorageAppendWrite100KB",bs)
	}
}

func BenchmarkStorageAppendWrite1MB(t *testing.B) {
	b:=byte(48)
	bs:=[]byte{}
	for i:=0;i<1024*1024-1;i++{
		bs=append(bs,b)
	}
	bs=append(bs,byte(10))
	storage:=newStorage(DefaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.AppendWrite("BenchmarkStorageAppendWrite1MB",bs)
	}
}
func BenchmarkStorageAppendWrite10MB(t *testing.B) {
	b:=byte(48)
	bs:=[]byte{}
	for i:=0;i<1024*1024*10-1;i++{
		bs=append(bs,b)
	}
	bs=append(bs,byte(10))
	storage:=newStorage(DefaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.AppendWrite("BenchmarkStorageAppendWrite10MB",bs)
	}
}

//go test -v -bench=.  -benchmem -benchtime=10s
//=== RUN   TestStorageSafeOverWrite10B
//--- PASS: TestStorageSafeOverWrite10B (0.01s)
//goos: darwin
//goarch: amd64
//pkg: hslam.com/mgit/Mort/raft
//BenchmarkStorageSafeOverWrite10B-4   	    3000	   5182259 ns/op	     698 B/op	      10 allocs/op
//BenchmarkStorageOverWrite10B-4       	    3000	   4433203 ns/op	     201 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite10B-4     	    3000	   3984780 ns/op	     202 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite1KB-4     	    3000	   3952492 ns/op	     200 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite10KB-4    	    3000	   4011653 ns/op	     201 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite100KB-4   	    3000	   4575498 ns/op	     201 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite1MB-4     	    2000	   6982571 ns/op	     200 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite10MB-4    	    1000	  19553136 ns/op	     202 B/op	       4 allocs/op
//PASS
//ok  	hslam.com/mgit/Mort/raft	117.785s