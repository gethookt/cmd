package wire

import (
	"encoding/json"
)

type Workflow struct {
	Jobs []Job `json:"jobs"`
}

type Job struct {
	ID      string   `json:"id"`
	Plugins []Plugin `json:"plugins"`
	Steps   []Step   `json:"steps"`
}

type Message struct {
	P []byte `json:"bytes,omitempty"`
}

func MakeMessage(p json.RawMessage) *Message {
	return &Message{P: p}
}

func (m *Message) Bytes() []byte { return m.P }

func (m *Message) Object() any {
	var o map[string]any
	if err := json.Unmarshal(m.P, &o); err != nil {
		panic("unexpected error: " + err.Error())
	}
	return o
}

type Generic = json.RawMessage

type Object map[string]Generic

type Step struct {
	Uses    string          `json:"uses"`
	ID      string          `json:"id,omitempty"`
	Desc    string          `json:"desc,omitempty"`
	With    json.RawMessage `json:"with"`
	Defer   string          `json:"defer,omitempty"`
	Timeout string          `json:"timeout,omitempty"`
	WaitFor string          `json:"wait_for,omitempty"`
}

type Plugin struct {
	Uses string          `json:"uses"`
	ID   string          `json:"id,omitempty"`
	With json.RawMessage `json:"with"`
}
