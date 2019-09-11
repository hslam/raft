package node

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"hslam.com/mgit/Mort/raft"
	"github.com/gorilla/mux"
	"hslam.com/mgit/Mort/rpc"
)

type Node struct {
	mu			sync.RWMutex
	host		string
	port		int
	rpc_port	int
	data_dir	string
	router		*mux.Router
	raft_node 	*raft.Node
	http_server	*http.Server
	http_client *http.Client
	db			*DB

}

func NewNode(data_dir string, host string, port ,rpc_port,raft_port int,peers []string) *Node {
	n := &Node{
		host:   	host,
		port:   	port,
		rpc_port:rpc_port,
		data_dir:   data_dir,
		db:     	newDB(),
		router: 	mux.NewRouter(),
	}
	n.http_client = &http.Client{Transport: &http.Transport{
		MaxIdleConnsPerHost: 1024,
	}}

	var err error
	n.raft_node, err = raft.NewNode(host,raft_port,n.data_dir,n.db)
	if err != nil {
		log.Fatal(err)
	}
	raft.SetLogLevel(0)
	n.raft_node.SetNode(peers)
	n.raft_node.RegisterCommand(&SetCommand{})
	n.raft_node.SetSnapshot(&Snapshot{})
	n.raft_node.SetCodec(&raft.ProtoCodec{})
	n.http_server = &http.Server{
		Addr:    fmt.Sprintf(":%d", n.port),
		Handler: n.router,
	}
	n.router.HandleFunc("/db/{key}", n.getHandler).Methods("GET")
	n.router.HandleFunc("/db/{key}", n.setHandler).Methods("POST")
	return n
}


func (n *Node) ListenAndServe() error {
	n.raft_node.Start()
	log.Println("Listening at:", n.uri())
	service:=new(Service)
	service.node=n
	server:= rpc.NewServer()
	server.RegisterName("S",service)
	server.EnableAsyncHandleWithSize(1024*256)
	rpc.SetLogLevel(99)
	go server.ListenAndServe("tcp", fmt.Sprintf(":%d", n.rpc_port))
	return n.http_server.ListenAndServe()
}

func (n *Node) uri() string {
	return fmt.Sprintf("http://%s:%d",  n.host, n.port)
}

func (n *Node) getHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	value := n.db.Get(vars["key"])
	w.Write([]byte(value))
}

func (n *Node) setHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	value := string(b)
	_, err = n.raft_node.Do(newSetCommand(vars["key"], value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
