// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package raft

func (entry *Entry) Encode(codec Codec, buf []byte) ([]byte, error) {
	return codec.Marshal(buf, entry)
}

func (entry *Entry) Decode(codec Codec, data []byte) {
	codec.Unmarshal(data, entry)
}
