// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

type Server struct {
	node *Node
	addr string
}

func newServer(node *Node, addr string) *Server {
	s := &Server{
		node: node,
		addr: addr,
	}
	return s
}
func (s *Server) listenAndServe() {
	go listenAndServe(s.addr, s.node)
}
