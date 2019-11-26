package main

import(
	"time"
	"hslam.com/git/x/raft"
	"math/rand"
)

type Client struct {
	node *raft.Node
	ctx *Context
	operation string
}
func (c *Client)Call()(int64,int64,bool){
	switch c.operation {
	case "set":
		c.node.Do(&Command{RandString(16)})
		return 0,0,true
	case "lease":
		if ok:=c.node.Lease();ok{
			value := c.ctx.Get()
			if value=="foobar"{
				return 0,0,true
			}
		}
	case "readindex":
		if ok:=c.node.ReadIndex();ok{
			value := c.ctx.Get()
			if value=="foobar"{
				return 0,0,true
			}
		}
	}
	return 0,0,false
}
func RandString(len int) string {
	r:= rand.New(rand.NewSource(time.Now().Unix()))
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}
