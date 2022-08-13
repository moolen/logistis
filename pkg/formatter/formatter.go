package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/moolen/logistis/pkg/store"
	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

func DiffEvent(ev *store.Event) (string, error) {
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

func FormatDiff(in string, lookAround int) string {
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
