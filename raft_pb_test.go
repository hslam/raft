// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"reflect"
	"testing"
)

func TestEntry(t *testing.T) {
	{
		var obj = Entry{Index: 1024, Term: 1024, CommandType: 64, Command: make([]byte, 512)}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy Entry
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = Entry{Index: 1024, Term: 1024, CommandType: 64, Command: make([]byte, 512)}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy Entry
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestRequestVoteRequest(t *testing.T) {
	{
		var obj = RequestVoteRequest{Term: 1024, CandidateId: "localhost:9001", LastLogIndex: 1024, LastLogTerm: 1024}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy RequestVoteRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = RequestVoteRequest{Term: 1024, CandidateId: "localhost:9001", LastLogIndex: 1024, LastLogTerm: 1024}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy RequestVoteRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestRequestVoteResponse(t *testing.T) {
	{
		var obj = RequestVoteResponse{Term: 1024, VoteGranted: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy RequestVoteResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = RequestVoteResponse{Term: 1024, VoteGranted: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy RequestVoteResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestAppendEntriesRequest(t *testing.T) {
	{
		var obj = AppendEntriesRequest{Term: 1024, LeaderId: "localhost:9001", PrevLogIndex: 1024, PrevLogTerm: 1024, LeaderCommit: 1024, Entries: []*Entry{{Index: 1024, Term: 1024, CommandType: 64, Command: make([]byte, 512)}}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy AppendEntriesRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = AppendEntriesRequest{Term: 1024, LeaderId: "localhost:9001", PrevLogIndex: 1024, PrevLogTerm: 1024, LeaderCommit: 1024, Entries: []*Entry{{Index: 1024, Term: 1024, CommandType: 64, Command: make([]byte, 512)}}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy AppendEntriesRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestAppendEntriesResponse(t *testing.T) {
	{
		var obj = AppendEntriesResponse{Term: 1024, Success: true, NextIndex: 1024}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy AppendEntriesResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = AppendEntriesResponse{Term: 1024, Success: true, NextIndex: 1024}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy AppendEntriesResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestInstallSnapshotRequest(t *testing.T) {
	{
		var obj = InstallSnapshotRequest{Term: 1024, LeaderId: "localhost:9001", LastIncludedIndex: 1024, LastIncludedTerm: 1024, Offset: 1024, Done: true, Data: make([]byte, 512)}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy InstallSnapshotRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = InstallSnapshotRequest{Term: 1024, LeaderId: "localhost:9001", LastIncludedIndex: 1024, LastIncludedTerm: 1024, Offset: 1024, Done: true, Data: make([]byte, 512)}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy InstallSnapshotRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestInstallSnapshotResponse(t *testing.T) {
	{
		var obj = InstallSnapshotResponse{Term: 1024, Offset: 1024, NextIndex: 1024}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy InstallSnapshotResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = InstallSnapshotResponse{Term: 1024, Offset: 1024, NextIndex: 1024}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy InstallSnapshotResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestNodeInfo(t *testing.T) {
	{
		var obj = NodeInfo{Address: "localhost:9001", NonVoting: true, Data: make([]byte, 512)}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy NodeInfo
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = NodeInfo{Address: "localhost:9001", NonVoting: true, Data: make([]byte, 512)}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy NodeInfo
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestConfigurationStorage(t *testing.T) {
	{
		var obj = ConfigurationStorage{Nodes: []*NodeInfo{{Address: "localhost:9001", NonVoting: true, Data: make([]byte, 512)}}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy ConfigurationStorage
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = ConfigurationStorage{Nodes: []*NodeInfo{{Address: "localhost:9001", NonVoting: true, Data: make([]byte, 512)}}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy ConfigurationStorage
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestAddPeerCommand(t *testing.T) {
	{
		var obj = AddPeerCommand{NodeInfo: &NodeInfo{Address: "localhost:9001", NonVoting: true, Data: make([]byte, 512)}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy AddPeerCommand
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = AddPeerCommand{NodeInfo: &NodeInfo{Address: "localhost:9001", NonVoting: true, Data: make([]byte, 512)}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy AddPeerCommand
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestRemovePeerCommand(t *testing.T) {
	{
		var obj = RemovePeerCommand{Address: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy RemovePeerCommand
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = RemovePeerCommand{Address: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy RemovePeerCommand
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestNoOperationCommand(t *testing.T) {
	{
		var obj = NoOperationCommand{}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy NoOperationCommand
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = NoOperationCommand{}
		var c = GOGOPBCodec{}
		var objCopy NoOperationCommand
		data, _ := obj.Marshal()
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestReconfigurationCommand(t *testing.T) {
	{
		var obj = ReconfigurationCommand{}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy ReconfigurationCommand
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = ReconfigurationCommand{}
		var c = GOGOPBCodec{}
		var objCopy ReconfigurationCommand
		data, _ := obj.Marshal()
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestQueryLeaderRequest(t *testing.T) {
	{
		var obj = QueryLeaderRequest{}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy QueryLeaderRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = QueryLeaderRequest{}
		var c = GOGOPBCodec{}
		var objCopy QueryLeaderRequest
		data, _ := obj.Marshal()
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestQueryLeaderResponse(t *testing.T) {
	{
		var obj = QueryLeaderResponse{Term: 1024, LeaderId: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy QueryLeaderResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = QueryLeaderResponse{Term: 1024, LeaderId: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy QueryLeaderResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestAddPeerRequest(t *testing.T) {
	{
		var obj = AddPeerRequest{Node: &NodeInfo{Address: "localhost:9001", NonVoting: true, Data: make([]byte, 512)}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy AddPeerRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = AddPeerRequest{Node: &NodeInfo{Address: "localhost:9001", NonVoting: true, Data: make([]byte, 512)}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy AddPeerRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestAddPeerResponse(t *testing.T) {
	{
		var obj = AddPeerResponse{Success: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy AddPeerResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = AddPeerResponse{Success: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy AddPeerResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestRemovePeerRequest(t *testing.T) {
	{
		var obj = RemovePeerRequest{Address: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy RemovePeerRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = RemovePeerRequest{Address: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy RemovePeerRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestRemovePeerResponse(t *testing.T) {
	{
		var obj = RemovePeerResponse{Success: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy RemovePeerResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = RemovePeerResponse{Success: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy RemovePeerResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}
