package wire // import "hookt.dev/cmd/pkg/plugin/builtin/event/wire"

import (
	"encoding/json"
	"log/slog"
	"time"

	"hookt.dev/cmd/pkg/proto/wire"
)

type Config struct {
	Sources         []string `json:"sources"`
	Timeout         string   `json:"timeout,omitempty"`
	InactiveTimeout string   `json:"inactive_timeout,omitempty"`
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

func (c Config) String() string {
	p, _ := json.Marshal(c)
	return string(p)
}

type Step struct {
	Match wire.Object `json:"match"`
	Pass  wire.Object `json:"pass,omitempty"`
	Fail  wire.Object `json:"fail,omitempty"`
}
