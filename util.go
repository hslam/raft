package raft

import (
	"time"
	"math/rand"
)

func  randomTime(t time.Time,maxRange time.Duration)time.Time{
	return t.Add(randomDurationTime(maxRange))
}



func  randomDurationTime(maxRange time.Duration)time.Duration{
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return maxRange*time.Duration((r.Intn(999)+1))/time.Duration(1000)
}


func  min(a ,b uint64)uint64{
	if a<b{
		return a
	}
	return b
}

func  max(a ,b uint64)uint64{
	if a>b{
		return a
	}
	return b
}

