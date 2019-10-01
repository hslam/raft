package node

import (
	"hslam.com/mgit/Mort/raft"
	"net/url"
	"strconv"
)



type Service struct {
	node *Node
}

func (s *Service)Set(req *Request, res *Response) error {
	if s.node.raft_node.IsLeader(){
		setCommand:=newSetCommand(req.Key, req.Value)
		_, err := s.node.raft_node.Do(newSetCommand(req.Key, req.Value))
		setCommandPool.Put(setCommand)
		if err==nil{
			res.Ok=true
			return nil
		}
		if err==raft.ErrNotLeader{
			leader:=s.node.raft_node.Leader()
			if leader!=""{
				res.Leader=leader
			}
		}
		return err
	}else {
		leader:=s.node.raft_node.Leader()
		if leader!=""{
			leader_url, err := url.Parse("http://" + leader)
			if err!=nil{
				panic(err)
			}
			port,err:=strconv.Atoi(leader_url.Port())
			if err!=nil{
				panic(err)
			}
			leader_host:=leader_url.Hostname() + ":" + strconv.Itoa(port-1000)
			return getRPCs().Call(leader_host,req,res)
		}
		return raft.ErrNotLeader
	}
	//setCommand:=newSetCommand(req.Key, req.Value)
	//_, err := s.node.raft_node.Do(newSetCommand(req.Key, req.Value))
	//setCommandPool.Put(setCommand)
	//if err==nil{
	//	res.Ok=true
	//	return nil
	//}
	//if err==raft.ErrNotLeader{
	//	leader:=s.node.raft_node.Leader()
	//	if leader!=""{
	//		res.Leader=leader
	//	}
	//}
	//return err
}
