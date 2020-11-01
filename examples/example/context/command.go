package context

func (c *Command) Type() int32 {
	return 1
}

func (c *Command) Do(context interface{}) (interface{}, error) {
	ctx := context.(*Context)
	ctx.Set(c.Data)
	return nil, nil
}
