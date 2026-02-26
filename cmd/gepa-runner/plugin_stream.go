package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/go-go-golems/go-go-gepa/pkg/jsbridge"
)

func newCommandEventSink(w io.Writer, enabled bool, command string) jsbridge.EventSink {
	if !enabled || w == nil {
		return nil
	}
	command = strings.TrimSpace(command)
	var mu sync.Mutex

	return func(event jsbridge.Event) {
		record := map[string]any{
			"kind":    "plugin_stream",
			"command": command,
			"event":   event,
		}
		blob, err := json.Marshal(record)
		if err != nil {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		_, _ = fmt.Fprintf(w, "stream-event %s\n", string(blob))
	}
}
