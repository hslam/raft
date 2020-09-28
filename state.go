// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

const (
	follower  = "Follower"
	candidate = "Candidate"
	leader    = "Leader"
)

type state interface {
	Start()
	Update() bool
	FixedUpdate()
	String() string
	StepDown() state
	NextState() state
}
