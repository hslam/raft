package raft

import (
	"sync"
	"errors"
	"os"
	"encoding/json"
	"path"
	"io/ioutil"
)

const (
	DefaultDataDir = "default.raft"
	DefaultConfig = "config"
	DefaultTmpConfig = "config.tmp"
	DefaultLog = "log"
	DefaultSnapshot = "snapshot"
)

type Server struct {
	mu 			sync.RWMutex
	waitGroup 	sync.WaitGroup
	address		string
	data_dir    string
	rpcs		*RPCs
	nodes      map[string]*Node
	context 		*Context
	currentTerm		uint64
	votedFor   		string
	log				[]*Entry

	commitIndex		uint64
	lastApplied		uint64
}

func NewServer(address,data_dir string)(*Server,error){
	if address == "" {
		return nil, errors.New("address is empty")
	}
	if data_dir == "" {
		data_dir=DefaultDataDir
	}
	mkdir(data_dir)
	s := &Server{
		address:address,
		data_dir:data_dir,
		rpcs:newRPCs(address,[]string{}),
		context:newContext(),
		nodes:make(map[string]*Node),
	}
	listenAndServe(address)
	return s,nil
}

func (s *Server) ChangeRoleToLeader() {
	s.context.Change(1)
	s.context.Change(1)
}
func (s *Server) State()Role {
	return s.context.State()
}
func (s *Server) AddNode(address string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.nodes[address] != nil {
		return nil
	}
	if s.address != address {
		node := newNode(s,address)
		if s.context.State() == Leader {
			node.startHeartbeat()
		}
		s.nodes[address] = node
		//s.DispatchEvent(newEvent(AddPeerEventType, name, nil))
	}
	s.saveConfig()
	return nil
}

func (s *Server) saveConfig() {
	nodes := make([]string, len(s.nodes))
	i := 0
	for _, node := range s.nodes {
		nodes[i] = node.address
		i++
	}
	c := &Configuration{
		Nodes:       nodes,
	}
	b, _ := json.Marshal(c)
	config_path := path.Join(s.data_dir, DefaultConfig)
	tmp_config_path := path.Join(s.data_dir, DefaultTmpConfig)
	err := writeToFile(tmp_config_path, b)
	if err != nil {
		panic(err)
	}
	os.Rename(tmp_config_path, config_path)
}

func (s *Server) loadConfig() error {
	config_path := path.Join(s.data_dir, DefaultConfig)
	b, err := ioutil.ReadFile(config_path)
	if err != nil {
		return nil
	}
	config := &Configuration{}
	if err = json.Unmarshal(b, config); err != nil {
		return err
	}
	//s.log.updateCommitIndex(config.CommitIndex)
	return nil
}