package raft

import (
	"time"
	"math/rand"
	"encoding/binary"
)

func  randomTime(t time.Time,maxRange time.Duration)time.Time{
	return t.Add(randomDurationTime(maxRange))
}

func  randomDurationTime(maxRange time.Duration)time.Duration{
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return maxRange*time.Duration((r.Intn(999)+1))/time.Duration(1000)
}

func  minUint64(a ,b uint64)uint64{
	if a<b { return a }
	return b
}

func  maxInt(a ,b int)int{
	if a>b{ return a }
	return b
}

func  uint64ToBytes(v uint64)[]byte{
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func  bytesToUint64(b []byte)uint64{
	return uint64(binary.BigEndian.Uint64(b))
}

func  uint32ToBytes(v uint32)[]byte{
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

func  bytesToUint32(b []byte)uint32{
	return uint32(binary.BigEndian.Uint32(b))
}