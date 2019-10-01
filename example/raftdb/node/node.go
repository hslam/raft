package node

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"hslam.com/mgit/Mort/raft"
	"hslam.com/mgit/Mort/mux"
	"hslam.com/mgit/Mort/mux-x/proxy"
	"hslam.com/mgit/Mort/mux-x/render"
	"hslam.com/mgit/Mort/rpc"
	"net/url"
	"strconv"
)
var (
	setCommandPool			*sync.Pool
)
func init() {
	setCommandPool= &sync.Pool{
		New: func() interface{} {
			return &SetCommand{}
		},
	}
}
const LeaderPrefix  = "LEADER:"

type Node struct {
	mu			sync.RWMutex
	host		string
	port		int
	rpc_port	int
	data_dir	string
	router		*mux.Router
	raft_node 	*raft.Node
	http_server	*http.Server
	db			*DB

}

func NewNode(data_dir string, host string, port ,rpc_port,raft_port int,peers []string) *Node {
	n := &Node{
		host:   	host,
		port:   	port,
		rpc_port:rpc_port,
		data_dir:   data_dir,
		db:     	newDB(),
		router: 	mux.New(),
	}
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
	n.router.HandleFunc("/status", n.statusHandler).All()
	n.router.HandleFunc("/db/:key", n.handle(n.getHandler)).GET()
	n.router.HandleFunc("/db/:key",n.handle(n.setHandler)).POST()
	return n
}


func (n *Node) ListenAndServe() error {
	n.raft_node.Start()
	log.Println("Listening at:", n.uri())
	service:=new(Service)
	service.node=n
	server:= rpc.NewServer()
	server.RegisterName("S",service)
	server.EnableAsyncHandleWithSize(1024)
	rpc.SetLogLevel(99)
	go server.ListenAndServe("tcp", fmt.Sprintf(":%d", n.rpc_port))
	return n.http_server.ListenAndServe()
}

func (n *Node) uri() string {
	return fmt.Sprintf("http://%s:%d",  n.host, n.port)
}

func (n *Node) handle(hander http.HandlerFunc) http.HandlerFunc{
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	return func(w http.ResponseWriter, r *http.Request) {
		if n.raft_node.IsLeader(){
			hander(w,r)
		}else {
			leader:=n.raft_node.Leader()
			if leader!=""{
				leader_url, err := url.Parse("http://" + leader)
				if err!=nil{
					panic(err)
				}
				port,err:=strconv.Atoi(leader_url.Port())
				if err!=nil{
					panic(err)
				}
				leader_url.Host=leader_url.Hostname() + ":" + strconv.Itoa(port-2000)
				proxy.Proxy(w,r,leader_url.String()+r.URL.Path)
			}
		}
	}
}


func (n *Node) setHandler(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	params := n.router.Params(req)
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	value := string(b)
	setCommand:=newSetCommand(params["key"], value)
	_, err = n.raft_node.Do(setCommand)
	setCommandPool.Put(setCommand)
	if err != nil {
		if err==raft.ErrNotLeader{
			leader:=n.raft_node.Leader()
			if leader!=""{
				http.Error(w, LeaderPrefix+leader, http.StatusBadRequest)
				return
			}
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}
func (n *Node) getHandler(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	params := n.router.Params(req)
	value := n.db.Get(params["key"])
	w.Write([]byte(value))
}
func (n *Node) statusHandler(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	status:=&Status{
		IsLeader:n.raft_node.IsLeader(),
		Leader:n.raft_node.Leader(),
		Node:n.raft_node.GetAddress(),
		Peers:n.raft_node.GetPeers(),
	}
	render.WriteJson(w,req,http.StatusOK,status)
}
