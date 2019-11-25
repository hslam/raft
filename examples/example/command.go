package main

type Command struct {
	Key		string
	Value	string
}
func (c *Command) Type()int32{
	return 1
}
func (c *Command) UniqueID()string {
	return c.Key
}
func (c *Command) Do(context interface{})(interface{},error){
	ctx := context.(*Context)
	ctx.Set(c.Key, c.Value)
	return nil,nil
}