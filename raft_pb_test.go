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
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
}

func TestRequestVoteRequest(t *testing.T) {
	{
		var obj = RequestVoteRequest{Term: 1024, CandidateID: "localhost:9001", LastLogIndex: 1024, LastLogTerm: 1024}
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
		var obj = RequestVoteRequest{Term: 1024, CandidateID: "localhost:9001", LastLogIndex: 1024, LastLogTerm: 1024}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy RequestVoteRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
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
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
}

func TestAppendEntriesRequest(t *testing.T) {
	{
		var obj = AppendEntriesRequest{Term: 1024, LeaderID: "localhost:9001", PrevLogIndex: 1024, PrevLogTerm: 1024, LeaderCommit: 1024, Entries: []*Entry{{Index: 1024, Term: 1024, CommandType: 64, Command: make([]byte, 512)}}}
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
		var obj = AppendEntriesRequest{Term: 1024, LeaderID: "localhost:9001", PrevLogIndex: 1024, PrevLogTerm: 1024, LeaderCommit: 1024, Entries: []*Entry{{Index: 1024, Term: 1024, CommandType: 64, Command: make([]byte, 512)}}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy AppendEntriesRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
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
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
}

func TestInstallSnapshotRequest(t *testing.T) {
	{
		var obj = InstallSnapshotRequest{Term: 1024, LeaderID: "localhost:9001", LastIncludedIndex: 1024, LastIncludedTerm: 1024, Offset: 1024, Done: true, Data: make([]byte, 512)}
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
		var obj = InstallSnapshotRequest{Term: 1024, LeaderID: "localhost:9001", LastIncludedIndex: 1024, LastIncludedTerm: 1024, Offset: 1024, Done: true, Data: make([]byte, 512)}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy InstallSnapshotRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
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
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
}

func TestNodeInfo(t *testing.T) {
	{
		var obj = NodeInfo{Address: "localhost:9001", NonVoting: true}
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
		var obj = NodeInfo{Address: "localhost:9001", NonVoting: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy NodeInfo
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
}

func TestConfigurationStorage(t *testing.T) {
	{
		var obj = ConfigurationStorage{Nodes: []*NodeInfo{{Address: "localhost:9001", NonVoting: true}}}
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
		var obj = ConfigurationStorage{Nodes: []*NodeInfo{{Address: "localhost:9001", NonVoting: true}}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy ConfigurationStorage
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
}

func TestDefaultCommand(t *testing.T) {
	{
		var obj = DefaultCommand{Operation: 1024, NodeInfo: &NodeInfo{Address: "localhost:9001", NonVoting: true}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy DefaultCommand
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = DefaultCommand{Operation: 1024, NodeInfo: &NodeInfo{Address: "localhost:9001", NonVoting: true}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy DefaultCommand
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
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
		var obj = QueryLeaderResponse{Term: 1024, LeaderID: "localhost:9001"}
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
		var obj = QueryLeaderResponse{Term: 1024, LeaderID: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy QueryLeaderResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
}

func TestSetPeerRequest(t *testing.T) {
	{
		var obj = SetPeerRequest{Node: &NodeInfo{Address: "localhost:9001", NonVoting: true}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy SetPeerRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = SetPeerRequest{Node: &NodeInfo{Address: "localhost:9001", NonVoting: true}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy SetPeerRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
}

func TestSetPeerResponse(t *testing.T) {
	{
		var obj = SetPeerResponse{Success: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy SetPeerResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = SetPeerResponse{Success: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy SetPeerResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
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
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
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
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
}

func TestGetMetaRequest(t *testing.T) {
	{
		var obj = GetMetaRequest{}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy GetMetaRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = GetMetaRequest{}
		var c = GOGOPBCodec{}
		var objCopy GetMetaRequest
		data, _ := obj.Marshal()
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
}

func TestQueryGetMetaResponse(t *testing.T) {
	{
		var obj = GetMetaResponse{Meta: []byte("localhost:9001")}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy GetMetaResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = GetMetaResponse{Meta: []byte("localhost:9001")}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy GetMetaResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
}
