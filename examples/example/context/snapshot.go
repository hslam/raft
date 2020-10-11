package context

import (
	"io"
	"io/ioutil"
)

type Snapshot struct {
	context *Context
}

func NewSnapshot(context *Context) *Snapshot {
	return &Snapshot{context: context}
}

func (s *Snapshot) Save(w io.Writer) (int, error) {
	return w.Write([]byte(s.context.Get()))
}

func (s *Snapshot) Recover(r io.Reader) (int, error) {
	raw, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}
	s.context.Set(string(raw))
	return len(raw), err
}
