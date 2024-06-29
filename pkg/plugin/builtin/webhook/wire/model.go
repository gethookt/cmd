package wire // import "hookt.dev/cmd/pkg/plugin/builtin/webhook/wire"

import (
	"hookt.dev/cmd/pkg/proto/wire"
)

type Config struct {
	Endpoints wire.Object `json:"endpoints"`
	Do        *Handler    `json:"do"`
}

type Handler struct {
	Method  wire.Generic `json:"method"`
	Headers wire.Generic `json:"headers"`
	Body    wire.Generic `json:"body"`
}

type Step struct{}
