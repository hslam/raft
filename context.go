package raft

import "time"

type Context struct {
	s State
	ch chan int
}

func newContext()*Context{
	c:=new(Context)
	c.s = new(FollowerState)
	c.ch = make(chan int)
	go c.Update()
	return c
}
func (c *Context) Context() {
	c.s = nil
}

func (c *Context) SetState(state State) {
	c.s = state
}

func (c *Context) GetState()State {
	return c.s
}
func (c *Context) String()string {
	return c.s.String()
}
func (c *Context) Change(i int){
	c.ch<-i
}

func (c *Context)Update(){
	for {
		select {
		case i := <-c.ch:
			if i == 1 {
				c.SetState(c.s.NextState())
			} else if i == -1 {
				c.SetState(c.s.PreState())
			}else if i == 0 {
				return
			}
		default:
			c.s.Update()
		}
		time.Sleep(time.Millisecond)
	}
}
func T() {
	c := newContext()
	time.Sleep(time.Second * 3)
	c.Change(1)
	time.Sleep(time.Second * 3)
	c.Change(1)
	time.Sleep(time.Second * 3)
	c.Change(0)
	time.Sleep(time.Second * 3)
}
