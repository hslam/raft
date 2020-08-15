// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

const (
	Follower  = "Follower"
	Candidate = "Candidate"
	Leader    = "Leader"
)

type State interface {
	Start()
	Update() bool
	FixedUpdate()
	String() string
	StepDown() State
	NextState() State
}
