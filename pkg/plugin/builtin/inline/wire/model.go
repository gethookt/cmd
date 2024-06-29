package wire // import "hookt.dev/cmd/pkg/plugin/builtin/inline/wire"

import "encoding/json"

type Config struct {
	Publish Source `json:"publish"`
}

func (c Config) String() string {
	p, _ := json.Marshal(c)
	return string(p)
}

type Source struct {
	File string `json:"file,omitempty"`
}

type Step struct{}
