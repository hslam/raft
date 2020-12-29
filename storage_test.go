// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"os"
	"testing"
)

func TestStorageSafeOverWrite10B(t *testing.T) {
	b := []byte("123456789\n")
	storage := newStorage(defaultDataDir)
	defer os.RemoveAll(defaultDataDir)
	err := storage.SafeOverWrite("TestStorageSafeOverWrite10B", b)
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkStorageSafeOverWrite10B(t *testing.B) {
	b := []byte("123456789\n")
	storage := newStorage(defaultDataDir)
	defer os.RemoveAll(defaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.SafeOverWrite("BenchmarkStorageSafeOverWrite10B", b)
	}
}

func BenchmarkStorageOverWrite10B(t *testing.B) {
	b := []byte("123456789\n")
	storage := newStorage(defaultDataDir)
	defer os.RemoveAll(defaultDataDir)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.OverWrite("BenchmarkStorageOverWrite10B", b)
	}
}

func BenchmarkStorageSeekWrite10B(t *testing.B) {
	b := []byte("123456789\n")
	storage := newStorage(defaultDataDir)
	defer os.RemoveAll(defaultDataDir)
	storage.SeekWrite("BenchmarkStorageSeekWrite10B", 10, b)
	storage.SeekWrite("BenchmarkStorageSeekWrite10B", 20, b)
	storage.SeekWrite("BenchmarkStorageSeekWrite10B", 30, b)
	b = []byte("000000000\n")

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		storage.SeekWrite("BenchmarkStorageSeekWrite10B", 10, b)
	}
}
