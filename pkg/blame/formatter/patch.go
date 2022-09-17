package formatter

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/moolen/logistis/pkg/store"
	jsonpatch "github.com/snorwin/jsonpatch"
)

func formatPatch(t table.Writer, eventMap map[string][]*store.Event, debug bool) error {
	t.AppendHeader(table.Row{"key", "operation", "user", "time"})

	for key, events := range eventMap {
		for _, ev := range events {
			formattedText, err := formatEventPatch(ev)
			if err != nil {
				return fmt.Errorf("unable to format patch: %w", err)
			}
			t.AppendRow([]interface{}{
				key,
				ev.Operation,
				formatUser(ev.UserInfo),
				ev.Timestamp.Format(time.RFC3339),
				formattedText})
			t.AppendSeparator()
		}
	}
	t.Render()
	return nil
}

func formatEventPatch(ev *store.Event) (string, error) {
	old := make(map[string]interface{})
	new := make(map[string]interface{})
	leftObj := ev.Object
	rightObj := ev.OldObject
	if leftObj == nil {
		leftObj = []byte("{}")
	}
	if rightObj == nil {
		rightObj = []byte("{}")
	}
	json.Unmarshal(leftObj, &old)
	json.Unmarshal(rightObj, &new)
	patch, err := jsonpatch.CreateJSONPatch(old, new)

	out := ""
	for _, p := range patch.List() {
		out += fmt.Sprintf("%s %s\n", p.Operation, p.Path)
	}

	return out, err
}
