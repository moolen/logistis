package recorder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/moolen/logistis/pkg/store"
	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

func (a *Recorder) ListEvents(w http.ResponseWriter, r *http.Request) {
	a.Logger.Info("received list events")
	namespace := r.URL.Query().Get("namespace")
	events, err := a.store.List(namespace)
	if err != nil {
		a.Logger.Errorf("unable to list events: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jout, err := json.Marshal(events)
	if err != nil {
		a.Logger.Errorf("unable to marshal events: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("content-type", "application/json")
	w.Write(jout)
}

func (a *Recorder) DiffEvents(w http.ResponseWriter, r *http.Request) {
	a.Logger.Info("received list events")
	namespace := r.URL.Query().Get("namespace")
	eventMap, err := a.store.List(namespace)
	if err != nil {
		a.Logger.Errorf("unable to list events: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	diff, err := a.diffEvents(eventMap)
	if err != nil {
		a.Logger.Errorf("unable to diff events: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(diff)
}

func (a *Recorder) diffEvents(eventMap map[string][]*store.Event) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	for kind, events := range eventMap {
		fmt.Fprintf(buf, "=====================\n")
		fmt.Fprintf(buf, "%s\n", kind)
		fmt.Fprintf(buf, "=====================\n")

		for _, ev := range events {
			a.Logger.Infof("diffing event %s", string(ev.Object))
			a.Logger.Infof("diffing event %s", string(ev.OldObject))
			fmt.Fprintf(buf, "operation: %s\n", ev.Operation)
			fmt.Fprintf(buf, "time: %s\n", ev.Timestamp)
			fmt.Fprintf(buf, "userinfo: %s\n", ev.UserInfo)

			diffStr, err := diffEvent(ev)
			if err != nil {
				a.Logger.Error(err)
				continue
			}
			// do not show the whole diff,
			// only the added/removed and a bit of context.
			diffStr = formatDiff(diffStr, 4)

			fmt.Fprintln(buf, diffStr)
			fmt.Fprintf(buf, "---\n")
		}
	}

	return buf.Bytes(), nil
}

func diffEvent(ev *store.Event) (string, error) {
	differ := diff.New()
	d, err := differ.Compare(ev.OldObject, ev.Object)
	if err != nil {
		return "", err
	}

	var aJson map[string]interface{}
	err = json.Unmarshal(ev.OldObject, &aJson)
	if err != nil {
		return "", err
	}

	config := formatter.AsciiFormatterConfig{
		ShowArrayIndex: true,
		Coloring:       true,
	}

	formatter := formatter.NewAsciiFormatter(aJson, config)
	diffString, err := formatter.Format(d)
	if err != nil {
		return "", err
	}

	return diffString, nil
}

func formatDiff(in string, lookAround int) string {
	out := bytes.NewBuffer(nil)
	lines := strings.Split(in, "\n")
	mask := make(map[int]bool)

	// find lines starting with ansi escape
	for i, line := range lines {
		if strings.HasPrefix(line, "\x1b[") {
			lower := 0
			upper := len(lines) - 1
			if i-lookAround > 0 {
				lower = i - lookAround
			}
			if i+lookAround < len(lines)-1 {
				upper = i + lookAround
			}
			// mark range in mask
			for i := lower; i <= upper; i++ {
				mask[i] = true
			}
		}
	}

	idx := []int{}
	for k := range mask {
		idx = append(idx, k)
	}
	sort.Ints(idx)

	for i, j := range idx {
		// add "..."
		if i > 1 && idx[i-1] < j-1 {
			fmt.Fprintln(out, "[...]")
		}
		fmt.Fprintln(out, lines[j])
	}

	return out.String()
}
