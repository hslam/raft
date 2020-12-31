package context

import (
	"fmt"
	"github.com/hslam/code"
)

// Command represents a command.
type Command struct {
	Data string
}

// Size returns the size of the buffer required to represent the data when encoded.
func (d *Command) Size() int {
	var size uint64
	size += 11 + uint64(len(d.Data))
	return int(size)
}

// Marshal returns the encoded bytes.
func (d *Command) Marshal() ([]byte, error) {
	size := d.Size()
	buf := make([]byte, size)
	n, err := d.MarshalTo(buf[:size])
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}

// MarshalTo marshals into buf and returns the number of bytes.
func (d *Command) MarshalTo(buf []byte) (int, error) {
	var size = uint64(d.Size())
	if uint64(cap(buf)) >= size {
		buf = buf[:size]
	} else {
		return 0, fmt.Errorf("proto: buf is too short")
	}
	var offset uint64
	var n uint64
	if len(d.Data) > 0 {
		buf[offset] = 1<<3 | 2
		offset++
		n = code.EncodeString(buf[offset:], d.Data)
		offset += n
	}
	return int(offset), nil
}

// Unmarshal unmarshals from data.
func (d *Command) Unmarshal(data []byte) error {
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
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			n = code.DecodeString(data[offset:], &d.Data)
			offset += n
		}
	}
	return nil
}
