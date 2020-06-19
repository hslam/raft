package raft

import (
	"time"
)

type Proxy interface {
	QueryLeader(addr string) (term uint64, leaderId string, ok bool)
	AddPeer(addr string, info *NodeInfo) (success bool, ok bool)
	RemovePeer(addr string, Address string) (success bool, ok bool)
	HandleQueryLeader(req *QueryLeaderRequest, res *QueryLeaderResponse) error
	HandleAddPeer(req *AddPeerRequest, res *AddPeerResponse) error
	HandleRemovePeer(req *RemovePeerRequest, res *RemovePeerResponse) error
}

type proxy struct {
	node               *Node
	queryLeaderTimeout time.Duration
	addPeerTimeout     time.Duration
	removePeerTimeout  time.Duration
}

func newProxy(node *Node) Proxy {
	return &proxy{
		node:               node,
		queryLeaderTimeout: DefaultQueryLeaderTimeout,
		addPeerTimeout:     DefaultAddPeerTimeout,
		removePeerTimeout:  DefaultRemovePeerTimeout,
	}
}

func (p *proxy) QueryLeader(addr string) (term uint64, leaderId string, ok bool) {
	var req = &QueryLeaderRequest{}
	var ch = make(chan *QueryLeaderResponse, 1)
	var errCh = make(chan error, 1)
	go func(rpcs *RPCs, addr string, req *QueryLeaderRequest, ch chan *QueryLeaderResponse, errCh chan error) {
		var res = &QueryLeaderResponse{}
		if err := rpcs.Call(addr, rpcs.QueryLeaderServiceName(), req, res); err != nil {
			errCh <- err
		} else {
			ch <- res
		}
	}(p.node.rpcs, addr, req, ch, errCh)
	select {
	case res := <-ch:
		Tracef("raft.QueryLeader %s -> %s LeaderId %s", p.node.address, addr, res.LeaderId)
		return res.Term, res.LeaderId, true
	case err := <-errCh:
		Tracef("raft.QueryLeader %s -> %s error %s", p.node.address, addr, err.Error())
		return 0, "", false
	case <-time.After(p.addPeerTimeout):
		Tracef("raft.QueryLeader %s -> %s time out", p.node.address, addr)
	}
	return 0, "", false
}
func (p *proxy) AddPeer(addr string, info *NodeInfo) (success bool, ok bool) {
	var req = &AddPeerRequest{}
	req.Node = info
	var ch = make(chan *AddPeerResponse, 1)
	var errCh = make(chan error, 1)
	go func(rpcs *RPCs, addr string, req *AddPeerRequest, ch chan *AddPeerResponse, errCh chan error) {
		var res = &AddPeerResponse{}
		if err := rpcs.Call(addr, rpcs.AddPeerServiceName(), req, res); err != nil {
			errCh <- err
		} else {
			ch <- res
		}
	}(p.node.rpcs, addr, req, ch, errCh)
	select {
	case res := <-ch:
		Tracef("raft.AddPeer %s -> %s Success %t", p.node.address, addr, res.Success)
		return res.Success, true
	case err := <-errCh:
		Tracef("raft.AddPeer %s -> %s error %s", p.node.address, addr, err.Error())
		return false, false
	case <-time.After(p.addPeerTimeout):
		Tracef("raft.AddPeer %s -> %s time out", p.node.address, addr)
	}
	return false, false
}
func (p *proxy) RemovePeer(addr string, Address string) (success bool, ok bool) {
	var req = &RemovePeerRequest{}
	req.Address = Address
	var ch = make(chan *RemovePeerResponse, 1)
	var errCh = make(chan error, 1)
	go func(rpcs *RPCs, addr string, req *RemovePeerRequest, ch chan *RemovePeerResponse, errCh chan error) {
		var res = &RemovePeerResponse{}
		if err := rpcs.Call(addr, rpcs.RemovePeerServiceName(), req, res); err != nil {
			errCh <- err
		} else {
			ch <- res
		}
	}(p.node.rpcs, addr, req, ch, errCh)
	select {
	case res := <-ch:
		Tracef("raft.RemovePeer %s -> %s Success %t", p.node.address, addr, res.Success)
		return res.Success, true
	case err := <-errCh:
		Tracef("raft.RemovePeer %s -> %s error %s", p.node.address, addr, err.Error())
		return false, false
	case <-time.After(p.removePeerTimeout):
		Tracef("raft.RemovePeer %s -> %s time out", p.node.address, addr)
	}
	return false, false
}
func (p *proxy) HandleQueryLeader(req *QueryLeaderRequest, res *QueryLeaderResponse) error {
	if p.node.leader != "" {
		res.LeaderId = p.node.leader
		res.Term = p.node.term()
		return nil
	}
	peers := p.node.Peers()
	for i := 0; i < len(peers); i++ {
		term, leaderId, ok := p.QueryLeader(peers[i])
		if ok {
			res.LeaderId = leaderId
			res.Term = term
			return nil
		}
	}
	return ErrNotLeader
}
func (p *proxy) HandleAddPeer(req *AddPeerRequest, res *AddPeerResponse) error {
	if p.node.IsLeader() {
		_, err := p.node.do(NewAddPeerCommand(req.Node), DefaultCommandTimeout)
		if err == nil {
			_, err = p.node.do(NewReconfigurationCommand(), DefaultCommandTimeout)
			if err == nil {
				res.Success = true
				return nil
			}
			return err
		}
		return err
	} else {
		leader := p.node.Leader()
		if leader != "" {
			return p.node.rpcs.Call(leader, p.node.rpcs.AddPeerServiceName(), req, res)
		}
		peers := p.node.Peers()
		for i := 0; i < len(peers); i++ {
			_, leaderId, ok := p.QueryLeader(peers[i])
			if leaderId != "" && ok {
				return p.node.rpcs.Call(leaderId, p.node.rpcs.AddPeerServiceName(), req, res)
			}
		}
		return ErrNotLeader
	}
	return nil
}

func (p *proxy) HandleRemovePeer(req *RemovePeerRequest, res *RemovePeerResponse) error {
	if p.node.IsLeader() {
		_, err := p.node.do(NewRemovePeerCommand(req.Address), DefaultCommandTimeout)
		if err == nil {
			_, err = p.node.do(NewReconfigurationCommand(), DefaultCommandTimeout)
			if err == nil {
				res.Success = true
				return nil
			}
			return err
		}
		return err
	} else {
		leader := p.node.Leader()
		if leader != "" {
			return p.node.rpcs.Call(leader, p.node.rpcs.RemovePeerServiceName(), req, res)
		}
		peers := p.node.Peers()
		for i := 0; i < len(peers); i++ {
			_, leaderId, ok := p.QueryLeader(peers[i])
			if leaderId != "" && ok {
				return p.node.rpcs.Call(leaderId, p.node.rpcs.RemovePeerServiceName(), req, res)
			}
		}
		return ErrNotLeader
	}
	return nil
}
