// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"math/rand"
	"sync"
	"time"
)

type election struct {
	node                   *node
	once                   sync.Once
	onceDisabled           bool
	random                 bool
	startTime              time.Time
	defaultElectionTimeout time.Duration
	electionTimeout        time.Duration
}

func newElection(n *node, electionTimeout time.Duration) *election {
	e := &election{
		node:                   n,
		defaultElectionTimeout: electionTimeout,
		random:                 true,
	}
	return e
}

func (e *election) Reset() {
	e.startTime = time.Now()
	if e.onceDisabled {
		if e.random {
			e.electionTimeout = e.defaultElectionTimeout + randomDurationTime(e.defaultElectionTimeout)
		} else {
			e.electionTimeout = e.defaultElectionTimeout
		}
	}
	e.once.Do(func() {
		e.electionTimeout = defaultStartWait
	})
}

func (e *election) Random(random bool) {
	e.random = random
}

func (e *election) Timeout() bool {
	if e.startTime.Add(e.electionTimeout).Before(time.Now()) {
		e.onceDisabled = true
		return true
	}
	return false
}

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomDurationTime(maxRange time.Duration) time.Duration {
	return maxRange * time.Duration((random.Intn(900) + 100)) / time.Duration(1000)
}
