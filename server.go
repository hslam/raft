package raft


type Server struct {
	node					*Node
	address					string
}
func newServer(address string,node *Node, ) *Server {
	s:=&Server{
		node :					node,
		address :				address,
	}
	return s
}
func  (s *Server) listenAndServe(){
	go listenAndServe(s.address,s.node)
}