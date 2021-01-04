// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"fmt"
)

// Address returns a raft node address with the given host and port.
func Address(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func minUint64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func maxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func quickSort(a []uint64, left, right int) {
	length := len(a)
	if length <= 1 {
		return
	}
	if left == -999 {
		left = 0
	}
	if right == -999 {
		right = length - 1
	}
	var partitonIndex int
	if left < right {
		partitonIndex = partition(a, left, right)
		quickSort(a, left, partitonIndex-1)
		quickSort(a, partitonIndex+1, right)
	}
}

func partition(a []uint64, left, right int) int {
	pivot := left
	index := pivot + 1
	for i := index; i <= right; i++ {
		if a[i] < a[pivot] {
			swap(a, i, index)
			index++
		}
	}
	swap(a, pivot, index-1)
	return index - 1
}

func swap(a []uint64, i, j int) {
	a[i], a[j] = a[j], a[i]
}
