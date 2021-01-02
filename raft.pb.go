package raft

import (
	"fmt"
	"github.com/hslam/code"
)

// Entry represents a log entry.
type Entry struct {
	Index       uint64
	Term        uint64
	CommandType int32
	Command     []byte
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *Entry) Size() int {
	var size uint64
	size += 11
	size += 11
	size += 11
	size += 11 + uint64(len(d.Command))
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *Entry) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *Entry) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if d.Index != 0 {
		buf[offset] = 1<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.Index)
		offset += n
	}
	if d.Term != 0 {
		buf[offset] = 2<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.Term)
		offset += n
	}
	if d.CommandType != 0 {
		buf[offset] = 3<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], uint64(d.CommandType))
		offset += n
	}
	if len(d.Command) > 0 {
		buf[offset] = 4<<3 | 2
		offset++
		n = code.EncodeBytes(buf[offset:], d.Command)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *Entry) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.Index)
			offset += n
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Term", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.Term)
			offset += n
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommandType", wireType)
			}
			var CommandType uint64
			n = code.DecodeVarint(data[offset:], &CommandType)
			d.CommandType = int32(CommandType)
			offset += n
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Command", wireType)
			}
			n = code.DecodeBytes(data[offset:], &d.Command)
			offset += n
		}
	}
	return nil
}

// RequestVoteRequest represents a rpc request of requesting vote.
type RequestVoteRequest struct {
	Term         uint64
	CandidateID  string
	LastLogIndex uint64
	LastLogTerm  uint64
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *RequestVoteRequest) Size() int {
	var size uint64
	size += 11
	size += 11 + uint64(len(d.CandidateID))
	size += 11
	size += 11
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *RequestVoteRequest) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *RequestVoteRequest) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if d.Term != 0 {
		buf[offset] = 1<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.Term)
		offset += n
	}
	if len(d.CandidateID) > 0 {
		buf[offset] = 2<<3 | 2
		offset++
		n = code.EncodeString(buf[offset:], d.CandidateID)
		offset += n
	}
	if d.LastLogIndex != 0 {
		buf[offset] = 3<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.LastLogIndex)
		offset += n
	}
	if d.LastLogTerm != 0 {
		buf[offset] = 4<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.LastLogTerm)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *RequestVoteRequest) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Term", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.Term)
			offset += n
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CandidateID", wireType)
			}
			n = code.DecodeString(data[offset:], &d.CandidateID)
			offset += n
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastLogIndex", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.LastLogIndex)
			offset += n
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastLogTerm", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.LastLogTerm)
			offset += n
		}
	}
	return nil
}

// RequestVoteResponse represents a rpc response of requesting vote.
type RequestVoteResponse struct {
	Term        uint64
	VoteGranted bool
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *RequestVoteResponse) Size() int {
	var size uint64
	size += 11
	size += 11
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *RequestVoteResponse) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *RequestVoteResponse) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if d.Term != 0 {
		buf[offset] = 1<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.Term)
		offset += n
	}
	if d.VoteGranted {
		buf[offset] = 2<<3 | 0
		offset++
		n = code.EncodeBool(buf[offset:], d.VoteGranted)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *RequestVoteResponse) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Term", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.Term)
			offset += n
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field VoteGranted", wireType)
			}
			n = code.DecodeBool(data[offset:], &d.VoteGranted)
			offset += n
		}
	}
	return nil
}

// AppendEntriesRequest represents a rpc request of appending entries.
type AppendEntriesRequest struct {
	Term         uint64
	LeaderID     string
	PrevLogIndex uint64
	PrevLogTerm  uint64
	LeaderCommit uint64
	Entries      []*Entry
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *AppendEntriesRequest) Size() int {
	var size uint64
	size += 11
	size += 11 + uint64(len(d.LeaderID))
	size += 11
	size += 11
	size += 11
	for i := range d.Entries {
		if d.Entries[i] != nil {
			size += 11 + uint64((d.Entries[i]).Size())
		}
	}
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *AppendEntriesRequest) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *AppendEntriesRequest) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if d.Term != 0 {
		buf[offset] = 1<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.Term)
		offset += n
	}
	if len(d.LeaderID) > 0 {
		buf[offset] = 2<<3 | 2
		offset++
		n = code.EncodeString(buf[offset:], d.LeaderID)
		offset += n
	}
	if d.PrevLogIndex != 0 {
		buf[offset] = 3<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.PrevLogIndex)
		offset += n
	}
	if d.PrevLogTerm != 0 {
		buf[offset] = 4<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.PrevLogTerm)
		offset += n
	}
	if d.LeaderCommit != 0 {
		buf[offset] = 5<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.LeaderCommit)
		offset += n
	}
	if len(d.Entries) > 0 {
		tag := byte(6<<3 | 2)
		for _, v := range d.Entries {
			buf[offset] = tag
			offset++
			s, err := v.MarshalTo(buf[offset+10:])
			if err != nil {
				return 0, err
			}
			size := uint64(s)
			n = code.EncodeVarint(buf[offset:], size)
			copy(buf[offset+n:], buf[offset+10:offset+10+size])
			offset += n + size
		}
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *AppendEntriesRequest) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Term", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.Term)
			offset += n
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LeaderID", wireType)
			}
			n = code.DecodeString(data[offset:], &d.LeaderID)
			offset += n
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PrevLogIndex", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.PrevLogIndex)
			offset += n
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PrevLogTerm", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.PrevLogTerm)
			offset += n
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LeaderCommit", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.LeaderCommit)
			offset += n
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Entries", wireType)
			}
			for {
				var b []byte
				n = code.DecodeBytes(data[offset:], &b)
				var entry = &Entry{}
				entry.Unmarshal(b)
				d.Entries = append(d.Entries, entry)
				offset += n
				if offset < length {
					tmpTag := uint64(data[offset])
					if tmpTag == tag {
						offset++
						continue
					}
				}
				break
			}
		}
	}
	return nil
}

// AppendEntriesResponse represents a rpc response of appending entries.
type AppendEntriesResponse struct {
	Term      uint64
	Success   bool
	NextIndex uint64
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *AppendEntriesResponse) Size() int {
	var size uint64
	size += 11
	size += 11
	size += 11
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *AppendEntriesResponse) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *AppendEntriesResponse) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if d.Term != 0 {
		buf[offset] = 1<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.Term)
		offset += n
	}
	if d.Success {
		buf[offset] = 2<<3 | 0
		offset++
		n = code.EncodeBool(buf[offset:], d.Success)
		offset += n
	}
	if d.NextIndex != 0 {
		buf[offset] = 3<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.NextIndex)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *AppendEntriesResponse) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Term", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.Term)
			offset += n
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Success", wireType)
			}
			n = code.DecodeBool(data[offset:], &d.Success)
			offset += n
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NextIndex", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.NextIndex)
			offset += n
		}
	}
	return nil
}

// InstallSnapshotRequest represents a rpc request of installing snapshot.
type InstallSnapshotRequest struct {
	Term              uint64
	LeaderID          string
	LastIncludedIndex uint64
	LastIncludedTerm  uint64
	Offset            uint64
	Done              bool
	Data              []byte
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *InstallSnapshotRequest) Size() int {
	var size uint64
	size += 11
	size += 11 + uint64(len(d.LeaderID))
	size += 11
	size += 11
	size += 11
	size += 11
	size += 11 + uint64(len(d.Data))
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *InstallSnapshotRequest) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *InstallSnapshotRequest) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if d.Term != 0 {
		buf[offset] = 1<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.Term)
		offset += n
	}
	if len(d.LeaderID) > 0 {
		buf[offset] = 2<<3 | 2
		offset++
		n = code.EncodeString(buf[offset:], d.LeaderID)
		offset += n
	}
	if d.LastIncludedIndex != 0 {
		buf[offset] = 3<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.LastIncludedIndex)
		offset += n
	}
	if d.LastIncludedTerm != 0 {
		buf[offset] = 4<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.LastIncludedTerm)
		offset += n
	}
	if d.Offset != 0 {
		buf[offset] = 5<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.Offset)
		offset += n
	}
	if d.Done {
		buf[offset] = 6<<3 | 0
		offset++
		n = code.EncodeBool(buf[offset:], d.Done)
		offset += n
	}
	if len(d.Data) > 0 {
		buf[offset] = 7<<3 | 2
		offset++
		n = code.EncodeBytes(buf[offset:], d.Data)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *InstallSnapshotRequest) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Term", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.Term)
			offset += n
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LeaderID", wireType)
			}
			n = code.DecodeString(data[offset:], &d.LeaderID)
			offset += n
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastIncludedIndex", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.LastIncludedIndex)
			offset += n
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastIncludedTerm", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.LastIncludedTerm)
			offset += n
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Offset", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.Offset)
			offset += n
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Done", wireType)
			}
			n = code.DecodeBool(data[offset:], &d.Done)
			offset += n
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			n = code.DecodeBytes(data[offset:], &d.Data)
			offset += n
		}
	}
	return nil
}

// InstallSnapshotResponse represents a rpc response of installing snapshot.
type InstallSnapshotResponse struct {
	Term      uint64
	Offset    uint64
	NextIndex uint64
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *InstallSnapshotResponse) Size() int {
	var size uint64
	size += 11
	size += 11
	size += 11
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *InstallSnapshotResponse) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *InstallSnapshotResponse) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if d.Term != 0 {
		buf[offset] = 1<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.Term)
		offset += n
	}
	if d.Offset != 0 {
		buf[offset] = 2<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.Offset)
		offset += n
	}
	if d.NextIndex != 0 {
		buf[offset] = 3<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.NextIndex)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *InstallSnapshotResponse) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Term", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.Term)
			offset += n
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Offset", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.Offset)
			offset += n
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NextIndex", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.NextIndex)
			offset += n
		}
	}
	return nil
}

// NodeInfo represents a node info.
type NodeInfo struct {
	Address   string
	NonVoting bool
	Data      []byte
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *NodeInfo) Size() int {
	var size uint64
	size += 11 + uint64(len(d.Address))
	size += 11
	size += 11 + uint64(len(d.Data))
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *NodeInfo) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *NodeInfo) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if len(d.Address) > 0 {
		buf[offset] = 1<<3 | 2
		offset++
		n = code.EncodeString(buf[offset:], d.Address)
		offset += n
	}
	if d.NonVoting {
		buf[offset] = 2<<3 | 0
		offset++
		n = code.EncodeBool(buf[offset:], d.NonVoting)
		offset += n
	}
	if len(d.Data) > 0 {
		buf[offset] = 3<<3 | 2
		offset++
		n = code.EncodeBytes(buf[offset:], d.Data)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *NodeInfo) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			n = code.DecodeString(data[offset:], &d.Address)
			offset += n
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NonVoting", wireType)
			}
			n = code.DecodeBool(data[offset:], &d.NonVoting)
			offset += n
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			n = code.DecodeBytes(data[offset:], &d.Data)
			offset += n
		}
	}
	return nil
}

// ConfigurationStorage represents a configuration storage.
type ConfigurationStorage struct {
	Nodes []*NodeInfo
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *ConfigurationStorage) Size() int {
	var size uint64
	for i := range d.Nodes {
		if d.Nodes[i] != nil {
			size += 11 + uint64((d.Nodes[i]).Size())
		}
	}
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *ConfigurationStorage) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *ConfigurationStorage) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if len(d.Nodes) > 0 {
		tag := byte(1<<3 | 2)
		for _, v := range d.Nodes {
			buf[offset] = tag
			offset++
			s, err := v.MarshalTo(buf[offset+10:])
			if err != nil {
				return 0, err
			}
			size := uint64(s)
			n = code.EncodeVarint(buf[offset:], size)
			copy(buf[offset+n:], buf[offset+10:offset+10+size])
			offset += n + size
		}
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *ConfigurationStorage) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Nodes", wireType)
			}
			for {
				var b []byte
				n = code.DecodeBytes(data[offset:], &b)
				var nodeInfo = &NodeInfo{}
				nodeInfo.Unmarshal(b)
				d.Nodes = append(d.Nodes, nodeInfo)
				offset += n
				if offset < length {
					tmpTag := uint64(data[offset])
					if tmpTag == tag {
						offset++
						continue
					}
				}
				break
			}
		}
	}
	return nil
}

// AddPeerCommand represents a command of adding peer.
type AddPeerCommand struct {
	NodeInfo *NodeInfo
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *AddPeerCommand) Size() int {
	var size uint64
	size += 11 + uint64(d.NodeInfo.Size())
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *AddPeerCommand) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *AddPeerCommand) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if d.NodeInfo != nil {
		buf[offset] = byte(1<<3 | 2)
		offset++
		s, err := d.NodeInfo.MarshalTo(buf[offset+10:])
		if err != nil {
			return 0, err
		} else if s == 0 {
			return 0, nil
		}
		size := uint64(s)
		n = code.EncodeVarint(buf[offset:], size)
		copy(buf[offset+n:], buf[offset+10:offset+10+size])
		offset += n + size
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *AddPeerCommand) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NodeInfo", wireType)
			}
			var b []byte
			n = code.DecodeBytes(data[offset:], &b)
			if d.NodeInfo == nil {
				d.NodeInfo = &NodeInfo{}
			}
			d.NodeInfo.Unmarshal(b)
			offset += n
		}
	}
	return nil
}

// RemovePeerCommand represents a command of removing peer.
type RemovePeerCommand struct {
	Address string
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *RemovePeerCommand) Size() int {
	var size uint64
	size += 11 + uint64(len(d.Address))
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *RemovePeerCommand) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *RemovePeerCommand) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if len(d.Address) > 0 {
		buf[offset] = 1<<3 | 2
		offset++
		n = code.EncodeString(buf[offset:], d.Address)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *RemovePeerCommand) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			n = code.DecodeString(data[offset:], &d.Address)
			offset += n
		}
	}
	return nil
}

// NoOperationCommand represents a command of no operation.
type NoOperationCommand struct {
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *NoOperationCommand) Size() int {
	return 0
}

// Marshal returns the encoded bytes.
func (d *NoOperationCommand) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *NoOperationCommand) MarshalTo(buf []byte) (int, error) {
	return 0, nil
}

// Unmarshal unmarshals from data.
func (d *NoOperationCommand) Unmarshal(data []byte) error {
	return nil
}

// ReconfigurationCommand represents a command of reconfiguration.
type ReconfigurationCommand struct {
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *ReconfigurationCommand) Size() int {
	return 0
}

// Marshal returns the encoded bytes.
func (d *ReconfigurationCommand) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *ReconfigurationCommand) MarshalTo(buf []byte) (int, error) {
	return 0, nil
}

// Unmarshal unmarshals from data.
func (d *ReconfigurationCommand) Unmarshal(data []byte) error {
	return nil
}

// QueryLeaderRequest represents a rpc request of querying leader.
type QueryLeaderRequest struct {
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *QueryLeaderRequest) Size() int {
	return 0
}

// Marshal returns the encoded bytes.
func (d *QueryLeaderRequest) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *QueryLeaderRequest) MarshalTo(buf []byte) (int, error) {
	return 0, nil
}

// Unmarshal unmarshals from data.
func (d *QueryLeaderRequest) Unmarshal(data []byte) error {
	return nil
}

// QueryLeaderResponse represents a rpc response of querying leader.
type QueryLeaderResponse struct {
	Term     uint64
	LeaderID string
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *QueryLeaderResponse) Size() int {
	var size uint64
	size += 11
	size += 11 + uint64(len(d.LeaderID))
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *QueryLeaderResponse) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *QueryLeaderResponse) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if d.Term != 0 {
		buf[offset] = 1<<3 | 0
		offset++
		n = code.EncodeVarint(buf[offset:], d.Term)
		offset += n
	}
	if len(d.LeaderID) > 0 {
		buf[offset] = 2<<3 | 2
		offset++
		n = code.EncodeString(buf[offset:], d.LeaderID)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *QueryLeaderResponse) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Term", wireType)
			}
			n = code.DecodeVarint(data[offset:], &d.Term)
			offset += n
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LeaderID", wireType)
			}
			n = code.DecodeString(data[offset:], &d.LeaderID)
			offset += n
		}
	}
	return nil
}

// AddPeerRequest represents a rpc request of adding peer.
type AddPeerRequest struct {
	Node *NodeInfo
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *AddPeerRequest) Size() int {
	var size uint64
	size += 11 + uint64(d.Node.Size())
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *AddPeerRequest) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *AddPeerRequest) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if d.Node != nil {
		buf[offset] = byte(1<<3 | 2)
		offset++
		s, err := d.Node.MarshalTo(buf[offset+10:])
		if err != nil {
			return 0, err
		} else if s == 0 {
			return 0, nil
		}
		size := uint64(s)
		n = code.EncodeVarint(buf[offset:], size)
		copy(buf[offset+n:], buf[offset+10:offset+10+size])
		offset += n + size
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *AddPeerRequest) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Node", wireType)
			}
			var b []byte
			n = code.DecodeBytes(data[offset:], &b)
			if d.Node == nil {
				d.Node = &NodeInfo{}
			}
			d.Node.Unmarshal(b)
			offset += n
		}
	}
	return nil
}

// AddPeerResponse represents a rpc response of adding peer.
type AddPeerResponse struct {
	Success  bool
	LeaderID string
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *AddPeerResponse) Size() int {
	var size uint64
	size += 11
	size += 11 + uint64(len(d.LeaderID))
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *AddPeerResponse) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *AddPeerResponse) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if d.Success {
		buf[offset] = 1<<3 | 0
		offset++
		n = code.EncodeBool(buf[offset:], d.Success)
		offset += n
	}
	if len(d.LeaderID) > 0 {
		buf[offset] = 2<<3 | 2
		offset++
		n = code.EncodeString(buf[offset:], d.LeaderID)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *AddPeerResponse) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Success", wireType)
			}
			n = code.DecodeBool(data[offset:], &d.Success)
			offset += n
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LeaderID", wireType)
			}
			n = code.DecodeString(data[offset:], &d.LeaderID)
			offset += n
		}
	}
	return nil
}

// RemovePeerRequest represents a rpc request of removing peer.
type RemovePeerRequest struct {
	Address string
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *RemovePeerRequest) Size() int {
	var size uint64
	size += 11 + uint64(len(d.Address))
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *RemovePeerRequest) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *RemovePeerRequest) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if len(d.Address) > 0 {
		buf[offset] = 1<<3 | 2
		offset++
		n = code.EncodeString(buf[offset:], d.Address)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *RemovePeerRequest) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			n = code.DecodeString(data[offset:], &d.Address)
			offset += n
		}
	}
	return nil
}

// RemovePeerResponse represents a rpc response of removing peer.
type RemovePeerResponse struct {
	Success  bool
	LeaderID string
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *RemovePeerResponse) Size() int {
	var size uint64
	size += 11
	size += 11 + uint64(len(d.LeaderID))
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *RemovePeerResponse) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	return buf[:n], err
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *RemovePeerResponse) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if d.Success {
		buf[offset] = 1<<3 | 0
		offset++
		n = code.EncodeBool(buf[offset:], d.Success)
		offset += n
	}
	if len(d.LeaderID) > 0 {
		buf[offset] = 2<<3 | 2
		offset++
		n = code.EncodeString(buf[offset:], d.LeaderID)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *RemovePeerResponse) Unmarshal(data []byte) error {
	var length = uint64(len(data))
	var offset uint64
	var n uint64
	var tag uint64
	var fieldNumber int
	var wireType uint8
	for {
		if offset < length {
			tag = uint64(data[offset])
			offset++
		} else {
			break
		}
		fieldNumber = int(tag >> 3)
		wireType = uint8(tag & 0x7)
		switch fieldNumber {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Success", wireType)
			}
			n = code.DecodeBool(data[offset:], &d.Success)
			offset += n
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LeaderID", wireType)
			}
			n = code.DecodeString(data[offset:], &d.LeaderID)
			offset += n
		}
	}
	return nil
}
