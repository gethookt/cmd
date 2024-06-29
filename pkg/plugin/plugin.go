package plugin // import "hookt.dev/cmd/pkg/plugin

import (
	"context"

	"hookt.dev/cmd/pkg/proto"
)

type Interface interface {
	Name() string
	Plugin(context.Context, *proto.P) any
	Step(context.Context) any
}
