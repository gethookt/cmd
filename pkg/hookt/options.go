package hookt

import "hookt.dev/cmd/pkg/proto"

func WithProtoOptions(opts ...func(*proto.P)) func(*Engine) {
	return func(e *Engine) {
		e.p = e.p.With(opts...)
	}
}
