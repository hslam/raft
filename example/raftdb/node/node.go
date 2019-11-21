package node

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"hslam.com/git/x/raft"
	"hslam.com/git/x/mux"
	"hslam.com/git/x/handler/proxy"
	"hslam.com/git/x/handler/render"
	"hslam.com/git/x/rpc"
	"net/url"
	"strconv"
)

const (
	network = "tcp"
	codec = "pb"
	MaxConnsPerHost=8
	MaxIdleConnsPerHost=0
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
	mux			*mux.Mux
	render 		*render.Render
	raft_node 	*raft.Node
	http_server	*http.Server
	rpc_transport	*rpc.Transport
	db			*DB

}

func NewNode(data_dir string, host string, port ,rpc_port,raft_port int,peers []string,join string) *Node {
	n := &Node{
		host:   	host,
		port:   	port,
		rpc_port:rpc_port,
		data_dir:   data_dir,
		db:     	newDB(),
		mux: 		mux.New(),
		render:		render.NewRender(),
	}
	var err error
	nodes :=make([]*raft.NodeInfo,len(peers))
	for i,v:=range peers{
		nodes[i]=&raft.NodeInfo{Address:v,Data:nil}
	}
	n.raft_node, err = raft.NewNode(host,raft_port,n.data_dir,n.db,nodes)
	if err != nil {
		log.Fatal(err)
	}
	raft.SetLogLevel(0)
	n.raft_node.RegisterCommand(&SetCommand{})
	n.raft_node.SetSnapshot(&Snapshot{})
	n.raft_node.SetSyncType([][]int{
		{86400,1},
		{14400,1000},
		{3600,10000},
		{1800,50000},
		{900,200000},
		{60,5000000},
	})
	n.raft_node.SetCodec(&raft.ProtoCodec{})
	n.raft_node.GzipSnapshot()
	n.http_server = &http.Server{
		Addr:    fmt.Sprintf(":%d", n.port),
		Handler: n.mux,
	}
	n.mux.Group("/cluster", func(m *mux.Mux) {
		m.HandleFunc("/status", n.statusHandler).All()
		m.HandleFunc("/leader", n.leaderHandler).All()
		m.HandleFunc("/ready", n.readyHandler).All()
		m.HandleFunc("/address", n.addressHandler).All()
		m.HandleFunc("/isleader", n.isLeaderHandler).All()
		m.HandleFunc("/peers", n.peersHandler).All()
		m.HandleFunc("/nodes", n.nodesHandler).All()
	})
	n.mux.HandleFunc("/db/:key", n.leaderHandle(n.getHandler)).GET()
	n.mux.HandleFunc("/db/:key",n.leaderHandle(n.setHandler)).POST()
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
	//server.EnableAsyncHandleWithSize(1024*256)
	server.EnableMultiplexingWithSize(1024*256)
	server.EnableBatch()
	rpc.SetLogLevel(99)
	if n.rpc_transport==nil{
		n.InitRPCProxy(MaxConnsPerHost,MaxIdleConnsPerHost)
	}

	go server.ListenAndServe("tcp", fmt.Sprintf(":%d", n.rpc_port))
	return n.http_server.ListenAndServe()
}
func (n *Node) InitRPCProxy(MaxConnsPerHost int,MaxIdleConnsPerHost int){
	opts:=rpc.DefaultOptions()
	opts.SetRetry(false)
	opts.SetCompressType("gzip")
	opts.SetMultiplexing(true)
	opts.SetBatch(true)
	opts.SetBatchAsync(true)
	n.rpc_transport=rpc.NewTransport(MaxConnsPerHost,MaxIdleConnsPerHost,network,codec,opts)
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
	params := n.mux.Params(req)
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
	params := n.mux.Params(req)
	if ok:=n.raft_node.ReadIndex();ok{
		value := n.db.Get(params["key"])
		w.Write([]byte(value))
	}
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
func (n *Node) readyHandler(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
		}
	}()
	n.render.JSON(w,req,n.raft_node.Ready(),http.StatusOK)
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
