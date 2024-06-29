package wire

import (
	"encoding/json"
	"log/slog"

	"hookt.dev/cmd/pkg/errors"

	"sigs.k8s.io/yaml"
)

type generic map[string]json.RawMessage

// TODO: use OAPI3.1 / JSONSchema
func XParse(p []byte) (*Workflow, error) {
	type (
		state byte
		node  struct {
			kind state
			data generic
			out  any
		}
	)

	const (
		start state = iota << 1
		job
		plugin
		step
	)

	var (
		nodes = make([]node, 1)
		it    node
		w     = &Workflow{}
	)

	if err := yamlUnmarshal(p, &nodes[0].data); err != nil {
		return nil, err
	}

	for len(nodes) != 0 {
		it, nodes = nodes[0], nodes[1:]

		switch it.kind {
		case start:
			for k, p := range it.data {
				switch k {
				case "jobs":
					var jobs []generic

					if err := yamlUnmarshal(p, &jobs); err != nil {
						return nil, errors.New("failed to unmarshal jobs: %w", err)
					}

					w.Jobs = make([]Job, len(jobs))

					for i, p := range jobs {
						nodes = append(nodes, node{
							kind: job,
							data: p,
							out:  &w.Jobs[i],
						})
					}
				default:
					return nil, errors.New("unsupported key for workflow: %q", k)
				}
			}
		case job:
			job := it.out.(*Job)

			for k, p := range it.data {
				switch k {
				case "id":
					if err := yamlUnmarshal(p, &job.ID); err != nil {
						return nil, errors.New("failed to unmarshal job.id: %w", err)
					}
				case "plugins":
					var plugins []generic

					if err := yamlUnmarshal(p, &plugins); err != nil {
						return nil, errors.New("failed to unmarshal plugins: %w", err)
					}

					job.Plugins = make([]Plugin, len(plugins))

					for i, p := range plugins {
						nodes = append(nodes, node{
							kind: plugin,
							data: p,
							out:  &job.Plugins[i],
						})
					}
				case "steps":
					var steps []generic

					if err := yamlUnmarshal(p, &steps); err != nil {
						return nil, errors.New("failed to unmarshal steps: %w", err)
					}

					job.Steps = make([]Step, len(steps))

					for i, p := range steps {
						nodes = append(nodes, node{
							kind: step,
							data: p,
							out:  &job.Steps[i],
						})
					}
				default:
					return nil, errors.New("unsupported key for job: %q", k)
				}
			}
		case plugin:
			plugin := it.out.(*Plugin)
			p := it.data

			slog.Debug("wire: plugin",
				"step", plugin.Uses,
				"with", str(p),
			)

			if err := yamlUnmarshalg(p, plugin); err != nil {
				return nil, errors.New("failed to unmarshal plugin: %w", err)
			}

			var x generic

			if err := yamlUnmarshal(plugin.With, &x); err != nil {
				return nil, errors.New("failed to unmarshal %q plugin.with: %w", plugin.Uses, err)
			}

			if len(x) == 0 {
				return nil, errors.New("plugin.with is empty for %q", plugin.Uses)
			}
		case step:
			step := it.out.(*Step)
			p := it.data

			slog.Debug("wire: step",
				"step", step.Uses,
				"with", str(p),
			)

			if err := yamlUnmarshalg(p, step); err != nil {
				return nil, errors.New("failed to unmarshal step: %w", err)
			}

			var x generic

			if err := yamlUnmarshal(step.With, &x); err != nil {
				return nil, errors.New("failed to unmarshal %q step.with: %w", step.Uses, err)
			}

			if len(x) == 0 {
				return nil, errors.New("step.with is empty for %q", step.Uses)
			}
		}
	}

	return w, nil
}

func yamlUnmarshal(p []byte, v any) error {
	return yaml.Unmarshal(p, v, yaml.DisallowUnknownFields)
}

func str(v any) string {
	p, _ := json.Marshal(v)
	return string(p)
}

func yamlUnmarshalg(g generic, v any) error {
	p, err := json.Marshal(g)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(p, v, yaml.DisallowUnknownFields)
}
