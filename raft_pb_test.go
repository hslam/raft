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
	{
		var obj Entry
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
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
	{
		var obj RequestVoteRequest
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
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
	{
		var obj RequestVoteResponse
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
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
	{
		var obj AppendEntriesRequest
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
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
	{
		var obj AppendEntriesResponse
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
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
	{
		var obj InstallSnapshotRequest
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
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
	{
		var obj InstallSnapshotResponse
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}

func TestMember(t *testing.T) {
	{
		var obj = Member{Address: "localhost:9001", NonVoting: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy Member
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = Member{Address: "localhost:9001", NonVoting: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy Member
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
	{
		var obj Member
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}

func TestConfigurationStorage(t *testing.T) {
	{
		var obj = ConfigurationStorage{Members: []*Member{{Address: "localhost:9001", NonVoting: true}, {Address: "localhost:9002", NonVoting: true}}}
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
		var obj = ConfigurationStorage{Members: []*Member{{Address: "localhost:9001", NonVoting: true}, {Address: "localhost:9002", NonVoting: true}}}
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
	{
		var obj ConfigurationStorage
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}

func TestDefaultCommand(t *testing.T) {
	{
		var obj = DefaultCommand{Operation: 1024, Member: &Member{Address: "localhost:9001", NonVoting: true}}
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
		var obj = DefaultCommand{Operation: 1024, Member: &Member{Address: "localhost:9001", NonVoting: true}}
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
	{
		var obj DefaultCommand
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}

func TestGetLeaderRequest(t *testing.T) {
	{
		var obj = GetLeaderRequest{}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy GetLeaderRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = GetLeaderRequest{}
		var c = GOGOPBCodec{}
		var objCopy GetLeaderRequest
		data, _ := obj.Marshal()
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj GetLeaderRequest
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}

func TestGetLeaderResponse(t *testing.T) {
	{
		var obj = GetLeaderResponse{Term: 1024, LeaderID: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy GetLeaderResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = GetLeaderResponse{Term: 1024, LeaderID: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy GetLeaderResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
	{
		var obj GetLeaderResponse
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}

func TestAddMemberRequest(t *testing.T) {
	{
		var obj = AddMemberRequest{Member: &Member{Address: "localhost:9001", NonVoting: true}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy AddMemberRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = AddMemberRequest{Member: &Member{Address: "localhost:9001", NonVoting: true}}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy AddMemberRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
	{
		var obj AddMemberRequest
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}

func TestAddMemberResponse(t *testing.T) {
	{
		var obj = AddMemberResponse{Success: true, LeaderID: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy AddMemberResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = AddMemberResponse{Success: true, LeaderID: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy AddMemberResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
	{
		var obj AddMemberResponse
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}

func TestRemoveMemberRequest(t *testing.T) {
	{
		var obj = RemoveMemberRequest{Address: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy RemoveMemberRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = RemoveMemberRequest{Address: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy RemoveMemberRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
	{
		var obj RemoveMemberRequest
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}

func TestRemoveMemberResponse(t *testing.T) {
	{
		var obj = RemoveMemberResponse{Success: true, LeaderID: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy RemoveMemberResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = RemoveMemberResponse{Success: true, LeaderID: "localhost:9001"}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy RemoveMemberResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
	{
		var obj RemoveMemberResponse
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}

func TestSetMetaRequest(t *testing.T) {
	{
		var obj = SetMetaRequest{Meta: []byte("localhost:9001")}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy SetMetaRequest
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = SetMetaRequest{Meta: []byte("localhost:9001")}
		var c = GOGOPBCodec{}
		var objCopy SetMetaRequest
		data, _ := obj.Marshal()
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
	{
		var obj SetMetaRequest
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}

func TestSetMetaResponse(t *testing.T) {
	{
		var obj = SetMetaResponse{Success: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 10240)
		var objCopy SetMetaResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
	}
	{
		var obj = SetMetaResponse{Success: true}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy SetMetaResponse
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
		if !reflect.DeepEqual(obj, objCopy) {
			t.Error()
		}
		if n, err := obj.MarshalTo(nil); n > 0 || err == nil {
			t.Error("The err should not be nil")
		}
	}
	{
		var obj SetMetaResponse
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
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
	{
		var obj GetMetaRequest
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}

func TestGetMetaResponse(t *testing.T) {
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
	{
		var obj GetMetaResponse
		var c = GOGOPBCodec{}
		n := byte(reflect.TypeOf(&obj).Elem().NumField())
		for i := byte(1); i <= n; i++ {
			if err := c.Unmarshal([]byte{i<<3 | 7}, &obj); err == nil {
				t.Error()
			}
		}
	}
}
