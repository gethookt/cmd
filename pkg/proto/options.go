package proto

func WithPlugins(plugins ...Interface) func(*P) {
	return func(p *P) {
		for _, plugin := range plugins {
			p.m[plugin.Name()] = plugin
		}
	}
}

func WithTOptions(opts ...TOption) func(*P) {
	return func(p *P) {
		p.t.Options = append(p.t.Options, opts...)
	}
}
