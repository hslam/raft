package raft

func (entry *Entry) Encode(codec Codec, buf []byte) ([]byte, error) {
	return codec.Marshal(buf, entry)
}

func (entry *Entry) Decode(codec Codec, data []byte) {
	codec.Unmarshal(data, entry)
}
