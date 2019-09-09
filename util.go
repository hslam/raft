package raft

import (
	"time"
	"math/rand"
	"encoding/binary"
	"bytes"
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
func  maxUint64(a ,b uint64)uint64{
	if a>b{ return a }
	return b
}
func  uint64ToBytes(v uint64)[]byte{
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
func  bytesToUint64(b []byte)uint64{
	return binary.BigEndian.Uint64(b)
}

func  uint32ToBytes(v uint32)[]byte{
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

func  bytesToUint32(b []byte)uint32{
	return uint32(binary.BigEndian.Uint32(b))
}

func Int64ToBytes(v int64) []byte {
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, v)
	return buffer.Bytes()
}


func  int32ToBytes(v int32)[]byte{
	b_buf := new(bytes.Buffer)
	binary.Write(b_buf, binary.BigEndian,v)
	return b_buf.Bytes()
}

func  bytesToInt32(b []byte)int32{
	var v int32
	b_buf := bytes.NewBuffer(b)
	binary.Read(b_buf, binary.BigEndian, v)
	return v
}



func quickSort(a []uint64,left ,right int){
	length:=len(a)
	if length <= 1{
		return
	}
	if left==-999{
		left=0
	}
	if right==-999{
		right=length-1
	}
	var partitonIndex int
	if left<right{
		partitonIndex=partition(a,left,right)
		quickSort(a,left,partitonIndex-1)
		quickSort(a,partitonIndex+1,right)
	}
}
func partition(a []uint64,left,right int)  int{
	pivot:=left
	index:=pivot+1
	for i:=index;i<=right;i++{
		if a[i]< a[pivot]{
			swap(a,i,index)
			index++
		}
	}
	swap(a,pivot,index-1)
	return index-1
}
func swap(a []uint64,i,j int)  {
	a[i],a[j]=a[j],a[i]
}