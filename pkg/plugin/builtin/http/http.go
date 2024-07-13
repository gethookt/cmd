package http // import "hookt.dev/cmd/pkg/plugin/builtin/http"

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"hookt.dev/cmd/pkg/check"
	"hookt.dev/cmd/pkg/plugin/builtin/http/wire"
	"hookt.dev/cmd/pkg/proto"
	protowire "hookt.dev/cmd/pkg/proto/wire"
	"hookt.dev/cmd/pkg/trace"
)

type Plugin struct {
	wire.Config

	h http.Header
	p *proto.P
}

func (p *Plugin) Name() string {
	return "http"
}

func New(opts ...func(*Plugin)) *Plugin {
	p := &Plugin{}
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

func (p *Plugin) Init(_ context.Context, job *proto.Job) (err error) {
	slog.Debug("http: init",
		"config", p.Config,
	)

	p.h, err = wire.Headers(p.Config.Headers, p.p)
	if err != nil {
		return err
	}

	return nil
}

func (p *Plugin) Step(context.Context) any {
	return &Step{
		p: p,
		c: &http.Client{
			Timeout: p.Config.GetTimeout(),
		},
	}
}

type Step struct {
	wire.Step `json:",inline"`

	p *Plugin
	c *http.Client
}

func (s *Step) Run(ctx context.Context, _ *check.S) error {
	req, err := s.req(ctx, &s.Request)
	if err != nil {
		return err
	}

	ctx = group(ctx, "pass")

	pass, err := s.p.p.Patterns(ctx, s.Step.Response.Pass)
	if err != nil {
		return err
	}

	resp, err := s.c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	p, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	r := &Response{
		Status:  resp.StatusCode,
		Headers: make(map[string]string),
	}

	if err := json.Unmarshal(p, &r.Body); err != nil {
		return err
	}

	for k := range resp.Header {
		r.Headers[k] = resp.Header.Get(k)
	}

	ok, err := pass.Match(ctx, r.Object())
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("response did not match pass pattern")
	}

	return nil
}

func group(ctx context.Context, name string) context.Context {
	return trace.With(ctx, "pattern-group", name)
}

type Response struct {
	Status  int               `json:"status,omitempty"`
	Headers map[string]string `json:"header,omitempty"`
	Body    any               `json:"body,omitempty"`
}

func (r *Response) Object() map[string]any {
	return map[string]any{
		"status": r.Status,
		"header": r.Headers,
		"body":   r.Body,
	}
}

func (s *Step) req(ctx context.Context, raw *wire.Request) (*http.Request, error) {
	var (
		res     wire.Request
		obj     protowire.Object
		headers = raw.Headers
	)

	raw.Headers = nil

	p, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(p, &obj); err != nil {
		return nil, err
	}

	if err := s.p.p.Template(ctx, obj, &res); err != nil {
		return nil, err
	}

	var body io.Reader
	if res.Body != "" {
		body = strings.NewReader(res.Body)
	}

	h, err := wire.Headers(headers, s.p.p)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, res.Method, res.URL, body)
	if err != nil {
		return nil, err
	}

	for k := range s.p.h {
		req.Header.Set(k, s.p.h.Get(k))
	}

	for k := range h {
		req.Header.Set(k, h.Get(k))
	}

	return req, nil
}

func (s *Step) Stop(context.Context) {}
