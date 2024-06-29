package builtin // import "hookt.dev/cmd/pkg/plugin/builtin"

import (
	"hookt.dev/cmd/pkg/plugin"
	"hookt.dev/cmd/pkg/plugin/builtin/event"
	"hookt.dev/cmd/pkg/plugin/builtin/http"
	"hookt.dev/cmd/pkg/plugin/builtin/inline"
	"hookt.dev/cmd/pkg/plugin/builtin/nats"
	"hookt.dev/cmd/pkg/plugin/builtin/webhook"
)

func Plugins() []plugin.Interface {
	return []plugin.Interface{
		event.New(),
		inline.New(),
		http.New(),
		nats.New(),
		webhook.New(),
	}
}
