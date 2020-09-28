// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

type server struct {
	node *node
	addr string
}

func newServer(n *node, addr string) *server {
	s := &server{
		node: n,
		addr: addr,
	}
	return s
}
func (s *server) listenAndServe() {
	go listenAndServe(s.addr, s.node)
}
