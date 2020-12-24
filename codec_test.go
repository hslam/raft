// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

import (
	"testing"
)

func TestJSONCodec(t *testing.T) {
	type Object struct {
		A bool `json:"A" xml:"A"`
	}
	var obj = Object{A: true}
	var c = JSONCodec{}
	var buf = make([]byte, 512)
	var objCopy *Object
	data, _ := c.Marshal(buf, &obj)
	c.Unmarshal(data, &objCopy)
}

func TestXMLCodec(t *testing.T) {
	type Object struct {
		A bool `json:"A" xml:"A"`
	}
	var obj = Object{A: true}
	var c = XMLCodec{}
	var buf = make([]byte, 512)
	var objCopy *Object
	data, _ := c.Marshal(buf, &obj)
	c.Unmarshal(data, &objCopy)
}

func TestBYTESCodec(t *testing.T) {
	var obj = []byte{128, 8, 128, 8, 195, 245, 72, 64, 74, 216, 18, 77, 251, 33, 9, 64, 10, 72, 101, 108, 108, 111, 87, 111, 114, 108, 100, 1, 1, 255, 2, 1, 128, 1, 255}
	var c = BYTESCodec{}
	var buf = make([]byte, 512)
	var objCopy []byte
	data, _ := c.Marshal(buf, &obj)
	c.Unmarshal(data, &objCopy)
}

//codeObject is a test struct
type codeObject struct {
}

//Marshal marshals the Object into buf and returns the bytes.
func (o *codeObject) Marshal(buf []byte) ([]byte, error) {
	return nil, nil
}

//Unmarshal unmarshals the Object from buf and returns the number of bytes read (> 0).
func (o *codeObject) Unmarshal(data []byte) (uint64, error) {
	return 0, nil
}

func TestCODECodec(t *testing.T) {
	{
		var obj = codeObject{}
		var c = CODECodec{}
		var buf = make([]byte, 512)
		var objCopy codeObject
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
	}
	{
		type Object struct {
		}
		var obj = Object{}
		var c = CODECodec{}
		var buf = make([]byte, 512)
		var objCopy Object
		if _, err := c.Marshal(buf, &obj); err != ErrorCODE {
			t.Error(ErrorCODE)
		}
		if err := c.Unmarshal(nil, &objCopy); err != ErrorCODE {
			t.Error(ErrorCODE)
		}
	}
}

type gogopbObject struct {
	size int
}

func (m *gogopbObject) Size() (n int) {
	return m.size
}

func (m *gogopbObject) Marshal() (dAtA []byte, err error) {
	return
}

func (m *gogopbObject) MarshalTo(dAtA []byte) (int, error) {
	return 0, nil
}

func (m *gogopbObject) Unmarshal(dAtA []byte) error {
	return nil
}

func TestGOGOPBCodec(t *testing.T) {
	{
		var obj = gogopbObject{}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 512)
		var objCopy gogopbObject
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
	}
	{
		var obj = gogopbObject{size: 128}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 0)
		var objCopy gogopbObject
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
	}
	{
		type Object struct {
		}
		var obj = Object{}
		var c = GOGOPBCodec{}
		var buf = make([]byte, 512)
		var objCopy Object
		if _, err := c.Marshal(buf, &obj); err != ErrorGOGOPB {
			t.Error(ErrorGOGOPB)
		}
		if err := c.Unmarshal(nil, &objCopy); err != ErrorGOGOPB {
			t.Error(ErrorGOGOPB)
		}
	}
}

type msgpObject struct {
}

// MarshalMsg implements msgp.Marshaler
func (z *msgpObject) MarshalMsg(b []byte) (o []byte, err error) {
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *msgpObject) UnmarshalMsg(bts []byte) (o []byte, err error) {
	return
}

func TestMSGPCodec(t *testing.T) {
	{
		var obj = msgpObject{}
		var c = MSGPCodec{}
		var buf = make([]byte, 512)
		var objCopy msgpObject
		data, _ := c.Marshal(buf, &obj)
		c.Unmarshal(data, &objCopy)
	}
	{
		type Object struct {
		}
		var obj = Object{}
		var c = MSGPCodec{}
		var buf = make([]byte, 512)
		var objCopy Object
		if _, err := c.Marshal(buf, &obj); err != ErrorMSGP {
			t.Error(ErrorMSGP)
		}
		if err := c.Unmarshal(nil, &objCopy); err != ErrorMSGP {
			t.Error(ErrorMSGP)
		}
	}
}
