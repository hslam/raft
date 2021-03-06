// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"testing"
)

func TestSort(t *testing.T) {
	arr := []uint64{1, 3, 2, 8, 24, 23, 22, 94, 56, 54, 73, 24, 19, 93, 34, 74, 52}
	quickSort(arr, -999, -999)
	for i := 0; i < len(arr)-1; i++ {
		if arr[i] > arr[i+1] {
			t.Error(arr[i], arr[i+1])
		}
	}
	quickSort(arr[:0], -999, -999)
}

func TestMinUint64(t *testing.T) {
	if minUint64(1, 2) != 1 {
		t.Error()
	}
}

func TestMaxUint64(t *testing.T) {
	if maxUint64(1, 2) != 2 {
		t.Error()
	}
}
