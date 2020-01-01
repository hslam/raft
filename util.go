package raft

import (
	"time"
	"math/rand"
	"strconv"
	"fmt"
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

func FormatName(v uint64) (s string) {
	getString:=func(n int,char string) (s string) {
		if n<1{return}
		for i:=1;i<=n;i++{s+=char}
		return
	}
	str:=strconv.FormatUint(v, 10)
	return fmt.Sprintf( "%s%s",getString(20-len(str),"0"),str)
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