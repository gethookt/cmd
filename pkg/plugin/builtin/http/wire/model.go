package wire // import "hookt.dev/cmd/pkg/plugin/builtin/http/wire"

import (
	"hookt.dev/cmd/pkg/proto/wire"
)

type Config struct {
	Headers wire.Object `json:"headers,omitempty"`
}

type Step struct {
	Request  *Request  `json:"request"`
	Response *Response `json:"response,omitempty"`
}

type Request struct {
	Method  string      `json:"method,omitempty"`
	URL     string      `json:"url"`
	Headers wire.Object `json:"headers,omitempty"`
	Body    string      `json:"body,omitempty"`
}

type Response struct {
	Status  int         `json:"status,omitempty"`
	Headers wire.Object `json:"headers,omitempty"`
	Body    wire.Object `json:"body,omitempty"`
}
