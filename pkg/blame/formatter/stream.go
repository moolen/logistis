package formatter

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/moolen/logistis/pkg/store"
	authenticationv1 "k8s.io/api/authentication/v1"
)

type Format string

const (
	DiffFormat  Format = "diff"
	PatchFormat Format = "patch"
)

func FormatStream(formatType Format, eventMap map[string][]*store.Event, debug bool) error {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	switch formatType {
	case DiffFormat, "":
		return formatDiff(t, eventMap, debug)
	case PatchFormat:
		return formatPatch(t, eventMap, debug)
	}
	return fmt.Errorf("unknown format type %q", formatType)
}

func formatUser(userInfo authenticationv1.UserInfo) string {
	out := userInfo.Username + "\n"
	if len(userInfo.Groups) > 0 {
		out += "groups:\n"
		for _, g := range userInfo.Groups {
			out += "- " + g + "\n"
		}
	}
	if len(userInfo.Extra) > 0 {
		out += "extra:\n"
		for k, vals := range userInfo.Extra {
			out += fmt.Sprintf("%s=%s\n", k, strings.Join(vals, ","))
		}

	}
	return out
}
