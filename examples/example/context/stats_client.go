package context

import (
	"github.com/hslam/raft"
	"math/rand"
)

type Client struct {
	Node      *raft.Node
	Ctx       *Context
	Operation string
}

func (c *Client) Call() (int64, int64, bool) {
	switch c.Operation {
	case "set":
		c.Node.Do(&Command{RandString(16)})
		return 0, 0, true
	case "lease":
		if ok := c.Node.Lease(); ok {
			value := c.Ctx.Get()
			if value == "foobar" {
				return 0, 0, true
			}
		}
	case "readindex":
		if ok := c.Node.ReadIndex(); ok {
			value := c.Ctx.Get()
			if value == "foobar" {
				return 0, 0, true
			}
		}
	}
	return 0, 0, false
}
func RandString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := rand.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}
