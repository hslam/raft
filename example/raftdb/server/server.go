package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"hslam.com/mgit/Mort/raft"
	"github.com/gorilla/mux"
	"hslam.com/mgit/Mort/raft/example/raftdb/command"
	"hslam.com/mgit/Mort/raft/example/raftdb/db"
)


type Server struct {
	mu			sync.RWMutex
	name		string
	host		string
	port		int
	path		string
	router		*mux.Router
	raftNode 	*raft.Node
	httpServer	*http.Server
	db			*db.DB
}

func New(path string, host string, port ,raft_port int) *Server {
	s := &Server{
		name:fmt.Sprintf(":%d", raft_port),
		host:   host,
		port:   port,
		path:   path,
		db:     db.New(),
		router: mux.NewRouter(),
	}
	return s
}

func (s *Server) connectionString() string {
	return fmt.Sprintf("http://%s:%d", s.host, s.port)
}

func (s *Server) ListenAndServe(leader string,peers []string) error {
	var err error
	s.raftNode, err = raft.NewNode(s.name, s.path,  s.db)
	if err != nil {
		log.Fatal(err)
	}
	s.raftNode.SetNode(peers)
	s.raftNode.Register(&command.WriteCommand{})
	s.raftNode.Start()
	if leader != "" {
		log.Println("Attempting to join leader:", leader)
	} else {
		log.Println("Recovered from log")
	}
	log.Println("Initializing HTTP server")

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.router,
	}

	s.router.HandleFunc("/db/{key}", s.readHandler).Methods("GET")
	s.router.HandleFunc("/db/{key}", s.writeHandler).Methods("POST")
	log.Println("Listening at:", s.connectionString())

	return s.httpServer.ListenAndServe()
}

func (s *Server) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.router.HandleFunc(pattern, handler)
}


func (s *Server) readHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	value := s.db.Get(vars["key"])
	w.Write([]byte(value))
}

func (s *Server) writeHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	value := string(b)
	_, err = s.raftNode.Do(command.NewWriteCommand(vars["key"], value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
