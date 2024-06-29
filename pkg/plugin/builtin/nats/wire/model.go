package wire // import "hookt.dev/cmd/pkg/plugin/builtin/nats/wire"

import (
	"hookt.dev/cmd/pkg/proto/wire"
)

type Config struct {
	Credentials string        `json:"credentials"`
	Subscribe   *Subscription `json:"subscribe"`
}

type Subscription struct {
	Subject string `json:"subject"`
}

type Step struct {
	Publish *Message `json:"publish"`
}

type Message struct {
	Subject  string      `json:"subject"`
	Encoding string      `json:"encoding,omitempty"`
	Data     wire.Object `json:"data"`
}
