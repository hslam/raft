// Copyright (c) 2019 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package context

func (c *Command) Type() int32 {
	return 1
}

func (c *Command) Do(context interface{}) (interface{}, error) {
	ctx := context.(*Context)
	ctx.Set(c.Data)
	return nil, nil
}
