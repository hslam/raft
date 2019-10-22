package raft

import (
	"sync"
	"time"
)

type ReadIndex struct {
	mu 						sync.Mutex
	node					*Node
	stop					chan bool
	finish 					chan bool
	readChan 				chan chan bool
	m 						map[uint64][]chan bool
	id 						uint64
	work 					bool
}

func newReadIndex(node *Node) *ReadIndex {
	r:=&ReadIndex{
		node:					node,
		stop:					make(chan bool,1),
		finish:					make(chan bool,1),
		readChan: 				make(chan chan bool,DefaultMaxConcurrencyRead),
		m: 						make(map[uint64][]chan bool),
		work:true,
	}
	go r.run()
	return r
}

func (r *ReadIndex) Read()(ok bool){
	var ch=make(chan bool,1)
	r.readChan<-ch
	select {
	case ok=<-ch:
	case <-time.After(DefaultCommandTimeout):
		ok=false
	}
	return
}

func (r *ReadIndex) reply(id uint64,success bool){
	defer func() {if err := recover(); err != nil {}}()
	r.mu.Lock()
	defer r.mu.Unlock()
	if _,ok:=r.m[id];ok{
		if len(r.m[id])>0{
			for _,ch:=range r.m[id]{
				ch<-success
			}
		}
	}
	delete(r.m,id)
}
func (r *ReadIndex) Update() {
	if r.work{
		r.work=false
		func(){
			defer func() {if err := recover(); err != nil {}}()
			defer func() {r.work=true}()
			r.mu.Lock()
			defer r.mu.Unlock()
			if _,ok:=r.m[r.id];ok{
				if len(r.m[r.id])>0{
					go func(node *Node,id uint64) {
						noOperationCommand:=newNoOperationCommand()
						if ok, _ := node.do(noOperationCommand,DefaultCommandTimeout);ok!=nil{
							r.reply(id,true)
							return
						}
						r.reply(id,false)
					}(r.node,r.id)
				}
			}
			if r.node.isLeader(){
				r.id+=1
			}
		}()
	}
}

func (r *ReadIndex) run() {
	go func() {
		for ch := range r.readChan {
			func(){
				r.mu.Lock()
				defer r.mu.Unlock()
				if _,ok:=r.m[r.id];!ok{
					r.m[r.id]=[]chan bool{}
				}
				r.m[r.id]=append(r.m[r.id],ch)
			}()
		}
	}()
	select {
	case <-r.stop:
		close(r.stop)
		r.stop=nil
		goto endfor
	}
endfor:
	r.finish<-true
}

func (r *ReadIndex)Stop()  {
	if r.stop==nil{
		return
	}
	r.stop<-true
	select {
	case <-r.finish:
		close(r.finish)
		r.finish=nil
	}
}