package raft

import (
	"github.com/golang/protobuf/proto"
	gogoproto "github.com/gogo/protobuf/proto"
	"encoding/json"
	"encoding/xml"
	"encoding/gob"
	"bytes"
)

type Codec interface {
	Encode(v interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) (error)
}

type JsonCodec struct{
}

func (c *JsonCodec) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *JsonCodec) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}


//type ProtoCodec struct{
//}
//
//func (c *ProtoCodec) Encode(v interface{}) ([]byte, error) {
//	return proto.Marshal(v.(proto.Message))
//}
//
//func (c *ProtoCodec) Decode(data []byte, v interface{}) error {
//	return proto.Unmarshal(data, v.(proto.Message))
//}


type XmlCodec struct{
}

func (c *XmlCodec) Encode(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

func (c *XmlCodec) Decode(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}


type GobCodec struct{
}

func (c *GobCodec) Encode(v interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	err :=  gob.NewEncoder(&buffer).Encode(v)
	if err!=nil{
		return nil,err
	}
	return buffer.Bytes(),nil
}

func (c *GobCodec) Decode(data []byte, v interface{}) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}

type BytesCodec struct{
}

func (c *BytesCodec) Encode(v interface{}) ([]byte, error) {
	return *v.(*[]byte),nil
}

func (c *BytesCodec) Decode(data []byte, v interface{}) error {
	*v.(*[]byte)=data
	return nil
}

type ProtoCodec struct{
}

func (c *ProtoCodec) Encode(v interface{}) ([]byte, error) {
	return gogoproto.Marshal(v.(proto.Message))
}

func (c *ProtoCodec) Decode(data []byte, v interface{}) error {
	return gogoproto.Unmarshal(data, v.(proto.Message))
}

type gen interface {
	Marshal(buf []byte) ([]byte, error)
	Unmarshal(buf []byte) (uint64, error)
}

type GencodeCodec struct{
	Buffer []byte
}

func (c *GencodeCodec) Encode(v interface{}) ([]byte, error) {
	return v.(gen).Marshal(c.Buffer)
}

func (c *GencodeCodec) Decode(data []byte, v interface{}) error {
	_,err:=v.(gen).Unmarshal(data)
	return err
}

type msgpack interface {
	MarshalMsg(buf []byte) ([]byte, error)
	UnmarshalMsg(bts []byte) (o []byte, err error)
}

type MsgpCodec struct{
	Buffer []byte
}

func (c *MsgpCodec) Encode(v interface{}) ([]byte, error) {
	return v.(msgpack).MarshalMsg(c.Buffer)
}

func (c *MsgpCodec) Decode(data []byte, v interface{}) error {
	_,err:=v.(msgpack).UnmarshalMsg(data)
	return err
}
