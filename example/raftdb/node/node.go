package node

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"hslam.com/mgit/Mort/raft"
	"hslam.com/mgit/Mort/mux"
	"hslam.com/mgit/Mort/handler/proxy"
	"hslam.com/mgit/Mort/handler/render"
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
	render 		*render.Render
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
		render:		render.NewRender(),
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
	n.raft_node.SetSnapshotSyncType(raft.EveryMinute)
	n.raft_node.SetCodec(&raft.ProtoCodec{})
	n.http_server = &http.Server{
		Addr:    fmt.Sprintf(":%d", n.port),
		Handler: n.router,
	}
	n.router.Group("/cluster", func(router *mux.Router) {
		router.HandleFunc("/status", n.statusHandler).All()
		router.HandleFunc("/leader", n.leaderHandler).All()
		router.HandleFunc("/address", n.addressHandler).All()
		router.HandleFunc("/isleader", n.isLeaderHandler).All()
		router.HandleFunc("/peers", n.peersHandler).All()
		router.HandleFunc("/nodes", n.nodesHandler).All()
	})
	n.router.HandleFunc("/db/:key", n.leaderHandle(n.getHandler)).GET()
	n.router.HandleFunc("/db/:key",n.leaderHandle(n.setHandler)).POST()
	n.router.Once()
	return n
}


func (n *Node) ListenAndServe() error {
	n.raft_node.Start()
	log.Println("HTTP listening at:", n.uri())
	log.Println("RPC listening at:", fmt.Sprintf("%s:%d", n.host,n.rpc_port))
	service:=new(Service)
	service.node=n
	server:= rpc.NewServer()
	server.RegisterName("S",service)
	server.EnableAsyncHandleWithSize(1024*256)
	rpc.SetLogLevel(99)
	if len(rpcsPool)==0{
		InitRPCProxy(MaxConnsPerHost)
	}
	go server.ListenAndServe("tcp", fmt.Sprintf(":%d", n.rpc_port))
	return n.http_server.ListenAndServe()
}

func (n *Node) uri() string {
	return fmt.Sprintf("http://%s:%d",  n.host, n.port)
}

func (n *Node) leaderHandle(hander http.HandlerFunc) http.HandlerFunc{
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
		Node:n.raft_node.Address(),
		Peers:n.raft_node.Peers(),
	}
	n.render.JSON(w,req,status,http.StatusOK)
}
func (n *Node) leaderHandler(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	w.Write([]byte(n.raft_node.Leader()))
}
func (n *Node) isLeaderHandler(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	n.render.JSON(w,req,n.raft_node.IsLeader(),http.StatusOK)
}
func (n *Node) addressHandler(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	w.Write([]byte(n.raft_node.Address()))
}
func (n *Node) peersHandler(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	n.render.JSON(w,req,n.raft_node.Peers(),http.StatusOK)
}
func (n *Node) nodesHandler(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	nodes:=n.raft_node.Peers()
	nodes=append(nodes,n.raft_node.Address())
	n.render.JSON(w,req,nodes,http.StatusOK)
}
