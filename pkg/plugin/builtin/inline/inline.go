package inline // import "hookt.dev/cmd/pkg/plugin/builtin/inline"

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"

	"hookt.dev/cmd/pkg/plugin/builtin/inline/wire"
	"hookt.dev/cmd/pkg/proto"
	protowire "hookt.dev/cmd/pkg/proto/wire"

	"github.com/lmittmann/tint"
)

type Plugin struct {
	wire.Config

	p *proto.P
	c chan proto.Message
}

func (p *Plugin) Name() string {
	return "inline"
}

func New(opts ...func(*Plugin)) *Plugin {
	p := &Plugin{
		c: make(chan proto.Message),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Plugin) WithProto(q *proto.P) *Plugin {
	p.p = q
	return p
}

func (p *Plugin) Plugin(_ context.Context, q *proto.P) any {
	return p.WithProto(q)
}

func (p *Plugin) Init(context.Context, *proto.Job) error {
	slog.Debug("inline: init",
		"config", p.Config,
	)

	file, err := p.p.Evaluate(p.Config.Publish.File, nil)
	if err != nil {
		return err
	}

	slog.Debug("inline: opening file",
		"file", string(file),
	)

	f, err := os.Open(string(file))
	if err != nil {
		return err
	}

	go p.publish(f)

	return nil
}

func (p *Plugin) publish(f *os.File) {
	defer f.Close()

	dec := json.NewDecoder(f)

	for index := 0; ; index++ {
		var raw json.RawMessage

		err := dec.Decode(&raw)
		if isEOF(err) {
			return
		}
		if err != nil {
			slog.Error("inline: publish",
				"raw", string(raw),
				tint.Err(err),
			)

			return
		}

		switch raw[0] {
		case '{':
			slog.Debug("inline: publish",
				"bytes", len(raw),
			)

			p.c <- &protowire.Message{P: raw, I: index}
		case '[':
			var msgs []json.RawMessage

			if err := json.Unmarshal(raw, &msgs); err != nil {
				slog.Error("inline: publish",
					"raw", string(raw),
					tint.Err(err),
				)
				return
			}

			for i := 0; i < len(msgs); i, index = i+1, index+1 {
				slog.Debug("inline: publish",
					"bytes", len(msgs[i]),
				)

				p.c <- &protowire.Message{P: msgs[i], I: index}
			}
		default:
			err = errors.New("unexpected JSON input")

			slog.Error("inline: publish",
				"input", string(raw),
				tint.Err(err),
			)

			return
		}

	}
}

func (p *Plugin) Subscribe(context.Context) <-chan proto.Message {
	return p.c
}

func (p *Plugin) Step(context.Context) any {
	return &Step{p: p}
}

func isEOF(err error) bool {
	const eof = "unexpected end of JSON input"

	if errors.Is(err, io.EOF) {
		return true
	}

	if e := new(json.SyntaxError); errors.As(err, &e) && strings.Contains(e.Error(), eof) {
		return true
	}

	return false
}

type Step struct {
	wire.Step `json:",inline"`

	p *Plugin
}

func (s *Step) Stop() {}
