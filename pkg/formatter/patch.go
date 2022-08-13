package formatter

import (
	"encoding/json"
	"fmt"

	"github.com/moolen/logistis/pkg/store"
	jsonpatch "github.com/snorwin/jsonpatch"
)

func FormatPatch(ev *store.Event) (string, error) {
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
