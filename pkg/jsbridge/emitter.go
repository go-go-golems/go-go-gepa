package jsbridge

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"
)

type Event struct {
	Kind                     string `json:"kind"`
	Sequence                 int64  `json:"sequence"`
	TimestampMS              int64  `json:"timestamp_ms"`
	PluginMethod             string `json:"plugin_method,omitempty"`
	PluginID                 string `json:"plugin_id,omitempty"`
	PluginName               string `json:"plugin_name,omitempty"`
	PluginRegistryIdentifier string `json:"plugin_registry_identifier,omitempty"`
	Type                     string `json:"type,omitempty"`
	Level                    string `json:"level,omitempty"`
	Message                  string `json:"message,omitempty"`
	Data                     any    `json:"data,omitempty"`
	Payload                  any    `json:"payload,omitempty"`
}

type EventSink func(Event)

type EmitterOptions struct {
	PluginMethod             string
	PluginID                 string
	PluginName               string
	PluginRegistryIdentifier string
	Sink                     EventSink
}

type Emitter struct {
	options EmitterOptions
	seq     atomic.Int64
}

func NewEmitter(options EmitterOptions) *Emitter {
	return &Emitter{options: options}
}

func (e *Emitter) Emit(payload any) {
	if e == nil || e.options.Sink == nil {
		return
	}
	event := Event{
		Kind:                     "plugin_event",
		Sequence:                 e.seq.Add(1),
		TimestampMS:              time.Now().UnixMilli(),
		PluginMethod:             strings.TrimSpace(e.options.PluginMethod),
		PluginID:                 strings.TrimSpace(e.options.PluginID),
		PluginName:               strings.TrimSpace(e.options.PluginName),
		PluginRegistryIdentifier: strings.TrimSpace(e.options.PluginRegistryIdentifier),
		Payload:                  payload,
	}

	switch raw := payload.(type) {
	case string:
		if strings.TrimSpace(raw) != "" {
			event.Message = raw
		}
	case map[string]any:
		if eventType := toTrimmedString(raw["type"]); eventType != "" {
			event.Type = eventType
		}
		if level := toTrimmedString(raw["level"]); level != "" {
			event.Level = level
		}
		if message := toTrimmedString(raw["message"]); message != "" {
			event.Message = message
		}
		if data, ok := raw["data"]; ok {
			event.Data = data
		}
		if event.Data == nil {
			event.Data = raw
		}
	default:
		event.Data = raw
	}

	defer func() {
		_ = recover()
	}()
	e.options.Sink(event)
}

func toTrimmedString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
