package wire // import "hookt.dev/cmd/pkg/plugin/builtin/http/wire"

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"hookt.dev/cmd/pkg/proto"
	"hookt.dev/cmd/pkg/proto/wire"
)

type Config struct {
	Timeout string      `json:"timeout,omitempty"`
	Headers wire.Object `json:"headers,omitempty"`
}

func (c Config) GetTimeout() time.Duration {
	if c.Timeout == "" {
		return 0
	}
	d, err := time.ParseDuration(c.Timeout)
	if err != nil {
		slog.Warn("ignoring invalid timeout",
			"timeout", c.Timeout,
		)
		return 0
	}
	return d
}

func Headers(raw wire.Object, p *proto.P) (http.Header, error) {
	var m map[string]string
	if err := p.Template(context.TODO(), raw, &m); err != nil {
		return nil, err
	}
	h := http.Header{}
	for k, v := range m {
		h.Set(k, v)
	}
	return h, nil
}

type Step struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Request struct {
	Method  string      `json:"method,omitempty"`
	URL     string      `json:"url"`
	Headers wire.Object `json:"headers,omitempty"`
	Body    string      `json:"body,omitempty"`
}

type Response struct {
	Pass wire.Object `json:"pass,omitempty"`
}
