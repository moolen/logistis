package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/moolen/logistis/pkg/store"
	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

func formatDiff(t table.Writer, eventMap map[string][]*store.Event, debug bool) error {
	for key, events := range eventMap {
		for _, ev := range events {
			formattedText := ""
			df, err := formatEventDiff(ev)
			if err != nil {
				return fmt.Errorf("unable to diff events: %w", err)
			}
			formattedText = wrapDiffOutput(df, 5)

			t.AppendRow([]interface{}{
				fmt.Sprintf("%s | %s | %s | %s",
					strconv.Itoa(int(time.Since(ev.Timestamp).Minutes()))+"m",
					key,
					ev.Operation,
					formatUser(ev.UserInfo))})
			t.AppendSeparator()
			t.AppendRow([]interface{}{
				formattedText})
			t.AppendSeparator()
		}
	}
	t.Render()
	return nil
}

func formatEventDiff(ev *store.Event) (string, error) {
	leftObj := ev.OldObject
	rightObj := ev.Object

	if leftObj == nil {
		leftObj = []byte("{}")
	}
	if rightObj == nil {
		rightObj = []byte("{}")
	}

	differ := diff.New()
	d, err := differ.Compare(leftObj, rightObj)
	if err != nil {
		return "", fmt.Errorf("unable to compare objects: %w", err)
	}

	var aJson map[string]interface{}
	err = json.Unmarshal(leftObj, &aJson)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal object: %w", err)
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

func wrapDiffOutput(in string, padding int) string {
	out := bytes.NewBuffer(nil)
	lines := strings.Split(in, "\n")

	// lines are capped in width
	maxWidth := 80

	// mask stores the lines which are relevant to us
	// we use it to calculate the padding
	mask := make(map[int]bool)

	// find lines starting with ansi escape
	for i, line := range lines {
		if strings.HasPrefix(line, "\x1b[") {
			lower := 0
			upper := len(lines) - 1
			if i-padding > 0 {
				lower = i - padding
			}
			if i+padding < len(lines)-1 {
				upper = i + padding
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
		line := lines[j]
		if len(line) > maxWidth {
			line = line[0:maxWidth]
		}
		fmt.Fprintln(out, line)
	}

	return out.String()
}
