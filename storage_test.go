// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"testing"
)

func TestStorageSafeOverWrite10B(t *testing.T) {
	b := []byte("123456789\n")
	storage := newStorage(&Node{}, defaultDataDir)
	err := storage.SafeOverWrite("TestStorageSafeOverWrite10B", b)
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkStorageSafeOverWrite10B(t *testing.B) {
	b := []byte("123456789\n")
	storage := newStorage(&Node{}, defaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.SafeOverWrite("BenchmarkStorageSafeOverWrite10B", b)
	}
}

func BenchmarkStorageOverWrite10B(t *testing.B) {
	b := []byte("123456789\n")
	storage := newStorage(&Node{}, defaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.OverWrite("BenchmarkStorageOverWrite10B", b)
	}
}
func BenchmarkStorageAppendWrite10B(t *testing.B) {
	b := []byte("123456789\n")
	storage := newStorage(&Node{}, defaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.AppendWrite("BenchmarkStorageAppendWrite10B", b)
	}
}
func BenchmarkStorageSeekWrite10B(t *testing.B) {
	b := []byte("123456789\n")
	storage := newStorage(&Node{}, defaultDataDir)
	storage.AppendWrite("BenchmarkStorageSeekWrite10B", b)
	storage.AppendWrite("BenchmarkStorageSeekWrite10B", b)
	storage.AppendWrite("BenchmarkStorageSeekWrite10B", b)
	b = []byte("000000000\n")

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.SeekWrite("BenchmarkStorageSeekWrite10B", 10, b)
	}
}
func BenchmarkStorageAppendWrite1KB(t *testing.B) {
	b := byte(48)
	bs := []byte{}
	for i := 0; i < 1023; i++ {
		bs = append(bs, b)
	}
	bs = append(bs, byte(10))
	storage := newStorage(&Node{}, defaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.AppendWrite("BenchmarkStorageAppendWrite1KB", bs)
	}
}
func BenchmarkStorageAppendWrite10KB(t *testing.B) {
	b := byte(48)
	bs := []byte{}
	for i := 0; i < 1024*10-1; i++ {
		bs = append(bs, b)
	}
	bs = append(bs, byte(10))
	storage := newStorage(&Node{}, defaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.AppendWrite("BenchmarkStorageAppendWrite10KB", bs)
	}
}
func BenchmarkStorageAppendWrite100KB(t *testing.B) {
	b := byte(48)
	bs := []byte{}
	for i := 0; i < 1024*100-1; i++ {
		bs = append(bs, b)
	}
	bs = append(bs, byte(10))
	storage := newStorage(&Node{}, defaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.AppendWrite("BenchmarkStorageAppendWrite100KB", bs)
	}
}

func BenchmarkStorageAppendWrite1MB(t *testing.B) {
	b := byte(48)
	bs := []byte{}
	for i := 0; i < 1024*1024-1; i++ {
		bs = append(bs, b)
	}
	bs = append(bs, byte(10))
	storage := newStorage(&Node{}, defaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.AppendWrite("BenchmarkStorageAppendWrite1MB", bs)
	}
}
func BenchmarkStorageAppendWrite10MB(t *testing.B) {
	b := byte(48)
	bs := []byte{}
	for i := 0; i < 1024*1024*10-1; i++ {
		bs = append(bs, b)
	}
	bs = append(bs, byte(10))
	storage := newStorage(&Node{}, defaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.AppendWrite("BenchmarkStorageAppendWrite10MB", bs)
	}
}

//go test -v -bench=.  -benchmem -benchtime=1s
//=== RUN   TestStorageSafeOverWrite10B
//--- PASS: TestStorageSafeOverWrite10B (0.01s)
//goos: darwin
//goarch: amd64
//pkg: github.com/hslam/raft
//BenchmarkStorageSafeOverWrite10B-4   	     300	   5119095 ns/op	     714 B/op	      10 allocs/op
//BenchmarkStorageOverWrite10B-4       	     300	   4492027 ns/op	     201 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite10B-4     	     300	   4045003 ns/op	     203 B/op	       4 allocs/op
//BenchmarkStorageSeekWrite10B-4       	     300	   4051719 ns/op	     200 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite1KB-4     	     300	   4415973 ns/op	     201 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite10KB-4    	     300	   4189121 ns/op	     202 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite100KB-4   	     300	   4781797 ns/op	     200 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite1MB-4     	     200	   6140275 ns/op	     202 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite10MB-4    	     100	  38680691 ns/op	     223 B/op	       4 allocs/op

//without f.Sync()
//BenchmarkStorageSafeOverWrite10B-4   	    5000	    376955 ns/op	     697 B/op	      10 allocs/op
//BenchmarkStorageAppendWrite10B-4     	   20000	     77311 ns/op	     200 B/op	       4 allocs/op
//BenchmarkStorageSeekWrite10B-4       	   20000	     78145 ns/op	     200 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite1KB-4     	   20000	     72821 ns/op	     200 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite10KB-4    	   10000	    131661 ns/op	     200 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite100KB-4   	   10000	    148539 ns/op	     200 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite1MB-4     	    2000	    527443 ns/op	     200 B/op	       4 allocs/op
//BenchmarkStorageAppendWrite10MB-4    	     300	   5280905 ns/op	     200 B/op	       4 allocs/op
