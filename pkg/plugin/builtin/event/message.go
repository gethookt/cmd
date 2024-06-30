package event

import (
	"hookt.dev/cmd/pkg/proto"
)

type Indexer interface {
	Index() int
}

type WaitMessage interface {
	proto.Message
	Done(bool)
	Wait() bool
}

type Message struct {
	proto.Message

	done chan bool
}

func (m *Message) Done(ok bool) {
	m.done <- ok
}

func (m *Message) Wait() bool {
	return <-m.done
}

var _ proto.Message = (*Message)(nil)

func Wait(msg proto.Message) WaitMessage {
	m := &Message{
		Message: msg,
		done:    make(chan bool),
	}

	if idx, ok := msg.(Indexer); ok {
		return struct {
			WaitMessage
			Indexer
		}{
			m,
			idx,
		}
	}

	return m
}
