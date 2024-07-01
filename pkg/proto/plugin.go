package proto

import (
	"context"

	"hookt.dev/cmd/pkg/check"
)

type Interface interface {
	Name() string
	Plugin(context.Context, *P) any
	Step(context.Context) any
}

type Message interface {
	Bytes() []byte
	Object() any
}

type Publisher interface {
	Publish(context.Context) chan<- Message
}

type Subscriber interface {
	Subscribe(context.Context) <-chan Message
}

type Initializer interface {
	Init(context.Context, *Job) error
}

type Runner interface {
	Run(context.Context, *check.S) error
	Stop(ctx context.Context)
}
