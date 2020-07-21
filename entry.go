package raft

func (entry *Entry) Encode(codec Codec) ([]byte, error) {
	return codec.Encode(entry)
}

func (entry *Entry) Decode(codec Codec, data []byte) {
	codec.Decode(data, entry)
}
